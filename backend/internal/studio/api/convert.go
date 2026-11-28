package api

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	adminv1 "trivia.al/protocol/gen/go/trivia/admin/v1"

	"trivia.al/backend/internal/platform/switches"
)

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

func toProto(r Row, effective bool) *adminv1.Switch {
	scope := adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL
	if r.ShowID != "" {
		scope = adminv1.SwitchScope_SWITCH_SCOPE_SHOW
	}
	s := &adminv1.Switch{Key: string(r.Key), Scope: scope, ShowId: r.ShowID, Enabled: r.Enabled, Effective: effective, ChangedBy: r.ChangedBy}
	if !r.ChangedAt.IsZero() {
		s.ChangedAt = timestamppb.New(r.ChangedAt)
	}
	return s
}

func defaultProto(key switches.Key, effective bool) *adminv1.Switch {
	return &adminv1.Switch{Key: string(key), Scope: adminv1.SwitchScope_SWITCH_SCOPE_GLOBAL,
		Enabled: switches.Default(key), Effective: effective, IsDefault: true}
}

func showOr(rowShow, requested string) string {
	if rowShow != "" {
		return rowShow
	}
	return requested
}
