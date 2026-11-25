// Package api exposes studio back-office endpoints. SwitchService lets an authenticated
// operator flip feature/kill switches (chat, reactions, extra lives, sign-ups) from the
// studio console, with a mandatory reason and an audit record (ADR-0009, SPEC §6.10).
//
// It is served only on the internal admin listener, never on the public game ALB.
package api

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	adminv1 "github.com/tirivia/protocol/gen/go/tirivia/admin/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/admin/v1/adminv1connect"

	"github.com/tirivia/backend/internal/platform/switches"
)

// MaxReasonBytes bounds the audit reason.
const MaxReasonBytes = 500

// Row is one stored switch value. ShowID is empty for global rows.
type Row struct {
	Key       switches.Key
	ShowID    string
	Enabled   bool
	ChangedBy string
	ChangedAt time.Time
}

// Store persists switches. Set must write feature_switches and audit_log in one transaction.
type Store interface {
	Rows(ctx context.Context) ([]Row, error)
	Set(ctx context.Context, row Row, reason string) error
}

// ActorFunc returns the authenticated operator id (from the BFF's verified SSO session).
type ActorFunc func(ctx context.Context) (string, error)

// SwitchHandler implements adminv1connect.SwitchServiceHandler.
type SwitchHandler struct {
	adminv1connect.UnimplementedSwitchServiceHandler

	store    Store
	actor    ActorFunc
	live     *switches.Memory // hot-path snapshot refreshed after every change
	clockNow func() time.Time
}

// NewSwitchHandler wires the handler. live may be nil when no in-process reader exists.
func NewSwitchHandler(store Store, actor ActorFunc, live *switches.Memory) *SwitchHandler {
	return &SwitchHandler{store: store, actor: actor, live: live, clockNow: time.Now}
}

// ListSwitches returns every known switch with its stored rows and effective value.
func (h *SwitchHandler) ListSwitches(ctx context.Context, req *adminv1.ListSwitchesRequest) (*adminv1.ListSwitchesResponse, error) {
	if _, err := h.actor(ctx); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, "operator session required")
	}
	rows, err := h.store.Rows(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "switch store unavailable")
	}
	snap := snapshotOf(rows)
	res := &adminv1.ListSwitchesResponse{}
	for _, key := range switches.Keys() {
		stored := false
		for _, r := range rows {
			if r.Key == key {
				stored = true
				res.Switches = append(res.Switches, toProto(r, snap.Enabled(key, pick(r.ShowID, req.GetShowId())), false))
			}
		}
		if !stored {
			res.Switches = append(res.Switches, &adminv1.Switch{
				Key: string(key), Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL,
				Enabled: switches.Default(key), Effective: snap.Enabled(key, req.GetShowId()), IsDefault: true,
			})
		}
	}
	return res, nil
}

// SetSwitch validates, persists (with audit) and applies a switch change.
func (h *SwitchHandler) SetSwitch(ctx context.Context, req *adminv1.SetSwitchRequest) (*adminv1.SetSwitchResponse, error) {
	actor, err := h.actor(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, "operator session required")
	}
	key := switches.Key(req.GetKey())
	if !switches.Known(key) {
		return nil, connect.NewError(connect.CodeInvalidArgument, "unknown switch key")
	}
	showID := strings.TrimSpace(req.GetShowId())
	switch req.GetScope() {
	case adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL:
		if showID != "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, "show_id must be empty for global scope")
		}
	case adminv1.SwitchScope_SWITCH_SCOPE_SHOW:
		if showID == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, "show_id is required for show scope")
		}
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, "scope is required")
	}
	reason := strings.TrimSpace(req.GetReason())
	if reason == "" || len(reason) > MaxReasonBytes || !utf8.ValidString(reason) {
		return nil, connect.NewError(connect.CodeInvalidArgument, "a reason (1-500 bytes) is required for the audit log")
	}

	row := Row{Key: key, ShowID: showID, Enabled: req.GetEnabled(), ChangedBy: actor, ChangedAt: h.clockNow().UTC()}
	if err := h.store.Set(ctx, row, reason); err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "switch store unavailable")
	}

	rows, err := h.store.Rows(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "switch store unavailable")
	}
	snap := snapshotOf(rows)
	if h.live != nil {
		h.live.Store(snap) // other tasks pick it up via the Valkey cache refresher (TTL <= 5 s)
	}
	return &adminv1.SetSwitchResponse{Switch: toProto(row, snap.Enabled(key, showID), false)}, nil
}

func snapshotOf(rows []Row) switches.Snapshot {
	snap := switches.Snapshot{Global: map[switches.Key]bool{}, PerShow: map[string]map[switches.Key]bool{}}
	for _, r := range rows {
		if r.ShowID == "" {
			snap.Global[r.Key] = r.Enabled
			continue
		}
		if snap.PerShow[r.ShowID] == nil {
			snap.PerShow[r.ShowID] = map[switches.Key]bool{}
		}
		snap.PerShow[r.ShowID][r.Key] = r.Enabled
	}
	return snap
}

func toProto(r Row, effective, isDefault bool) *adminv1.Switch {
	scope := adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL
	if r.ShowID != "" {
		scope = adminv1.SwitchScope_SWITCH_SCOPE_SHOW
	}
	s := &adminv1.Switch{
		Key: string(r.Key), Scope: scope, ShowId: r.ShowID, Enabled: r.Enabled,
		Effective: effective, IsDefault: isDefault, ChangedBy: r.ChangedBy,
	}
	if !r.ChangedAt.IsZero() {
		s.ChangedAt = timestamppb.New(r.ChangedAt)
	}
	return s
}

func pick(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ErrNoActor is returned by ActorFunc implementations when no operator session is present.
var ErrNoActor = errors.New("studio: no operator session")
