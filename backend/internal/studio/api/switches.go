package api

import (
	"context"
	"time"

	"connectrpc.com/connect/v2"

	adminv1 "trivia.al/protocol/gen/go/trivia/admin/v1"
	"trivia.al/protocol/gen/go/trivia/admin/v1/adminv1connect"

	"trivia.al/backend/internal/platform/switches"
)

// SwitchHandler implements adminv1connect.SwitchServiceHandler.
type SwitchHandler struct {
	adminv1connect.UnimplementedSwitchServiceHandler
	store Store
	actor ActorFunc
	live  *switches.Memory
	now   func() time.Time
}

// NewSwitchHandler wires the handler; live may be nil.
func NewSwitchHandler(store Store, actor ActorFunc, live *switches.Memory) *SwitchHandler {
	return &SwitchHandler{store: store, actor: actor, live: live, now: time.Now}
}

var (
	errUnauth      = connect.NewError(connect.CodeUnauthenticated, "operator session required")
	errUnavailable = connect.NewError(connect.CodeUnavailable, "switch store unavailable")
)

// ListSwitches returns every known switch with stored rows and effective values.
func (h *SwitchHandler) ListSwitches(ctx context.Context, req *adminv1.ListSwitchesRequest) (*adminv1.ListSwitchesResponse, error) {
	if _, err := h.actor(ctx); err != nil {
		return nil, errUnauth
	}
	rows, err := h.store.Rows(ctx)
	if err != nil {
		return nil, errUnavailable
	}
	snap := snapshotOf(rows)
	res := &adminv1.ListSwitchesResponse{}
	for _, key := range switches.Keys() {
		stored := false
		for _, r := range rows {
			if r.Key == key {
				stored = true
				res.Switches = append(res.Switches, toProto(r, snap.Enabled(key, showOr(r.ShowID, req.GetShowId()))))
			}
		}
		if !stored {
			res.Switches = append(res.Switches, defaultProto(key, snap.Enabled(key, req.GetShowId())))
		}
	}
	return res, nil
}

// SetSwitch validates, persists with audit, and applies a switch change.
func (h *SwitchHandler) SetSwitch(ctx context.Context, req *adminv1.SetSwitchRequest) (*adminv1.SetSwitchResponse, error) {
	actor, err := h.actor(ctx)
	if err != nil {
		return nil, errUnauth
	}
	key, showID, reason, err := validateSet(req)
	if err != nil {
		return nil, err
	}
	row := Row{Key: key, ShowID: showID, Enabled: req.GetEnabled(), ChangedBy: actor, ChangedAt: h.now().UTC()}
	if err := h.store.Set(ctx, row, reason); err != nil {
		return nil, errUnavailable
	}
	rows, err := h.store.Rows(ctx)
	if err != nil {
		return nil, errUnavailable
	}
	snap := snapshotOf(rows)
	if h.live != nil {
		h.live.Store(snap)
	}
	return &adminv1.SetSwitchResponse{Switch: toProto(row, snap.Enabled(key, showID))}, nil
}
