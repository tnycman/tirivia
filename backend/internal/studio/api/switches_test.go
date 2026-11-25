package api

import (
	"context"
	"errors"
	"sync"
	"testing"

	"connectrpc.com/connect/v2"

	adminv1 "github.com/tirivia/protocol/gen/go/tirivia/admin/v1"

	"github.com/tirivia/backend/internal/platform/switches"
)

// memStore stands in for Postgres (feature_switches + audit_log in one transaction).
type memStore struct {
	mu    sync.Mutex
	rows  map[string]Row
	audit []string
}

func newMemStore() *memStore { return &memStore{rows: map[string]Row{}} }

func (m *memStore) Rows(context.Context) ([]Row, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Row, 0, len(m.rows))
	for _, r := range m.rows {
		out = append(out, r)
	}
	return out, nil
}

func (m *memStore) Set(_ context.Context, r Row, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[string(r.Key)+"|"+r.ShowID] = r
	m.audit = append(m.audit, r.ChangedBy+":"+string(r.Key)+":"+reason)
	return nil
}

func operator(context.Context) (string, error) { return "op-1", nil }
func nobody(context.Context) (string, error)   { return "", ErrNoActor }

func code(err error) connect.Code {
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return connect.CodeUnknown
}

func TestChatToggleFromBackOfficeReachesGameServers(t *testing.T) {
	ctx := context.Background()
	store := newMemStore()
	live := switches.NewMemory(switches.Snapshot{})
	h := NewSwitchHandler(store, operator, live)

	if live.Enabled(ctx, switches.ChatEnabled, "show-1") {
		t.Fatal("chat must default to OFF")
	}

	// Operator turns chat on globally for a rehearsal.
	res, err := h.SetSwitch(ctx, &adminv1.SetSwitchRequest{
		Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL, Enabled: true, Reason: "rehearsal chat test",
	})
	if err != nil || !res.GetSwitch().GetEffective() {
		t.Fatalf("set on: %v / %v", res, err)
	}
	if !live.Enabled(ctx, switches.ChatEnabled, "show-1") {
		t.Fatal("game servers did not see chat ON")
	}

	// Production show decided without chat: show-scoped OFF overrides the global ON.
	if _, err := h.SetSwitch(ctx, &adminv1.SetSwitchRequest{
		Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_SHOW, ShowId: "show-live-1", Enabled: false, Reason: "launch without chat",
	}); err != nil {
		t.Fatal(err)
	}
	if live.Enabled(ctx, switches.ChatEnabled, "show-live-1") {
		t.Fatal("show-level OFF must win")
	}
	if len(store.audit) != 2 {
		t.Fatalf("every change must be audited, got %d records", len(store.audit))
	}

	list, err := h.ListSwitches(ctx, &adminv1.ListSwitchesRequest{ShowId: "show-live-1"})
	if err != nil || len(list.GetSwitches()) < len(switches.Keys()) {
		t.Fatalf("list: %v / %v", list, err)
	}
}

func TestSetSwitchRejects(t *testing.T) {
	ctx := context.Background()
	h := NewSwitchHandler(newMemStore(), operator, nil)
	cases := map[string]*adminv1.SetSwitchRequest{
		"unknown key":      {Key: "free.money", Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL, Reason: "x"},
		"no scope":         {Key: "chat.enabled", Reason: "x"},
		"show scope no id": {Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_SHOW, Reason: "x"},
		"global with id":   {Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL, ShowId: "s", Reason: "x"},
		"no reason":        {Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL},
	}
	for name, req := range cases {
		if _, err := h.SetSwitch(ctx, req); code(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: want InvalidArgument, got %v", name, err)
		}
	}
	anon := NewSwitchHandler(newMemStore(), nobody, nil)
	if _, err := anon.SetSwitch(ctx, &adminv1.SetSwitchRequest{Key: "chat.enabled", Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL, Reason: "x"}); code(err) != connect.CodeUnauthenticated {
		t.Errorf("no operator: want Unauthenticated, got %v", err)
	}
}
