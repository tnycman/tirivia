// Package keystest is the behaviour every appattest.Keys store must share.
package keystest

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tirivia/backend/internal/auth/attest/appattest"
)

// Run checks k: single insert per key, round trip, strictly increasing counters, one winner under races.
func Run(t *testing.T, k appattest.Keys) {
	t.Helper()
	ctx := context.Background()
	key := appattest.Key{ID: random(t, 32), PublicKey: random(t, 65), PrizeEligible: true, Receipt: []byte{1, 2}}
	if err := k.Insert(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err := k.Insert(ctx, key); !errors.Is(err, appattest.ErrKeyExists) {
		t.Fatalf("duplicate insert: %v", err)
	}
	got, ok, err := k.Get(ctx, key.ID)
	if err != nil || !ok || !bytes.Equal(got.PublicKey, key.PublicKey) || !got.PrizeEligible || got.Counter != 0 || !bytes.Equal(got.Receipt, key.Receipt) {
		t.Fatalf("round trip: %+v %v %v", got, ok, err)
	}
	if _, ok, err := k.Get(ctx, random(t, 32)); err != nil || ok {
		t.Fatalf("unknown key: %v %v", ok, err)
	}
	for _, step := range []struct {
		counter uint32
		want    bool
	}{{1, true}, {1, false}, {0, false}, {5, true}, {4, false}} {
		if won, err := k.AdvanceCounter(ctx, key.ID, step.counter); err != nil || won != step.want {
			t.Fatalf("advance to %d: %v %v", step.counter, won, err)
		}
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			if won, err := k.AdvanceCounter(ctx, key.ID, 6); err == nil && won {
				wins.Add(1)
			}
		})
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d concurrent advances to one counter won; want 1", wins.Load())
	}
}

func random(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}
