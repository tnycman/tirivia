package live

import (
	"context"
	"fmt"
	"maps"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/platform/livestate"
)

// Recount tallies the stored counted answers with cursored HSCAN, so no single command blocks the server; run it after the window closes.
func Recount(ctx context.Context, c valkey.Client, show string, q int) (map[string]int64, error) {
	out := map[string]int64{}
	for b := range livestate.TallyShards {
		answers, err := scanBucket(ctx, c, livestate.CountedKey(show, q, b))
		if err != nil {
			return nil, err
		}
		for _, opt := range answers {
			out[opt]++
		}
	}
	return out, nil
}

// scanBucket returns user → option; HSCAN may repeat an entry, so it is keyed by user.
func scanBucket(ctx context.Context, c valkey.Client, key string) (map[string]string, error) {
	answers := map[string]string{}
	for cursor := uint64(0); ; {
		e, err := c.Do(ctx, c.B().Hscan().Key(key).Cursor(cursor).Build()).AsScanEntry()
		if err != nil {
			return nil, err
		}
		for i := 0; i+1 < len(e.Elements); i += 2 {
			answers[e.Elements[i]] = e.Elements[i+1]
		}
		if cursor = e.Cursor; cursor == 0 {
			return answers, nil
		}
	}
}

// Verify checks the live answer tally against the recount (TS-16); extra lives are not answers, so they are left out.
func Verify(ctx context.Context, c valkey.Client, show string, q int) error {
	live, err := Sum(ctx, c, show, q)
	if err != nil {
		return err
	}
	delete(live, LifeOption)
	recount, err := Recount(ctx, c, show, q)
	if err != nil {
		return err
	}
	if !maps.Equal(live, recount) {
		return fmt.Errorf("live: tally %v differs from recount %v for show %s Q%d", live, recount, show, q)
	}
	return nil
}
