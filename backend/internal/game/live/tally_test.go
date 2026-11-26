package live

import (
	"context"
	"testing"
)

func TestTallyAcrossShards(t *testing.T) {
	c := client(t)
	ctx, show := context.Background(), id(t)
	a, b := NewTally(c, 3), NewTally(c, 41)
	for range 70 {
		a.Add(show, 1, "x")
	}
	for range 30 {
		b.Add(show, 1, "x")
		b.Add(show, 1, "y")
	}
	b.Add(show, 2, "x")
	for _, tl := range []*Tally{a, b} {
		if err := tl.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	a.Add(show, 1, "y")
	if err := a.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := Sum(ctx, c, show, 1)
	if err != nil || got["x"] != 100 || got["y"] != 31 || len(got) != 2 {
		t.Fatalf("Q1 totals: %v %v", got, err)
	}
	if q2, _ := Sum(ctx, c, show, 2); q2["x"] != 1 {
		t.Fatalf("Q2 totals: %v", q2)
	}
	if err := a.Flush(ctx); err != nil {
		t.Fatalf("empty flush: %v", err)
	}
}
