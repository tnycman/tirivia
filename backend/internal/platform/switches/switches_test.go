package switches

import (
	"context"
	"testing"
)

func TestResolution(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		snap Snapshot
		show string
		want bool
	}{
		{"no rows: chat defaults off", Snapshot{}, "s1", false},
		{"global on", Snapshot{Global: map[Key]bool{ChatEnabled: true}}, "s1", true},
		{"show off overrides global on", Snapshot{
			Global:  map[Key]bool{ChatEnabled: true},
			PerShow: map[string]map[Key]bool{"s1": {ChatEnabled: false}},
		}, "s1", false},
		{"show on overrides global off", Snapshot{
			Global:  map[Key]bool{ChatEnabled: false},
			PerShow: map[string]map[Key]bool{"s1": {ChatEnabled: true}},
		}, "s1", true},
		{"other show falls back to global", Snapshot{
			Global:  map[Key]bool{ChatEnabled: true},
			PerShow: map[string]map[Key]bool{"s1": {ChatEnabled: false}},
		}, "s2", true},
	}
	for _, tc := range cases {
		if got := NewMemory(tc.snap).Enabled(ctx, ChatEnabled, tc.show); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestLiveFlip(t *testing.T) {
	ctx := context.Background()
	m := NewMemory(Snapshot{Global: map[Key]bool{ChatEnabled: true}})
	if !m.Enabled(ctx, ChatEnabled, "s1") {
		t.Fatal("expected on")
	}
	m.Store(Snapshot{Global: map[Key]bool{ChatEnabled: false}}) // operator flips the kill switch
	if m.Enabled(ctx, ChatEnabled, "s1") {
		t.Fatal("expected off after flip")
	}
}
