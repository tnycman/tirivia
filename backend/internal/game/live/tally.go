package live

import (
	"context"
	"strconv"
	"sync"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/platform/livestate"
)

type tallyKey struct {
	show string
	q    int
}

// Tally pre-aggregates counts in memory; Flush writes them to this task's shard (SPEC §6.7).
type Tally struct {
	c       valkey.Client
	shard   int
	mu      sync.Mutex
	pending map[tallyKey]map[string]int64
}

// NewTally returns a Tally writing to shard (0 ≤ shard < livestate.TallyShards).
func NewTally(c valkey.Client, shard int) *Tally {
	return &Tally{c: c, shard: shard % livestate.TallyShards, pending: map[tallyKey]map[string]int64{}}
}

// Add counts one answer.
func (t *Tally) Add(show string, q int, option string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	k := tallyKey{show, q}
	if t.pending[k] == nil {
		t.pending[k] = map[string]int64{}
	}
	t.pending[k][option]++
}

// Flush pipelines one HINCRBY per pending option; failed counts are restored for the next flush.
func (t *Tally) Flush(ctx context.Context) error {
	t.mu.Lock()
	batch := t.pending
	t.pending = map[tallyKey]map[string]int64{}
	t.mu.Unlock()
	var cmds valkey.Commands
	var undo []func()
	for k, opts := range batch {
		for opt, n := range opts {
			cmds = append(cmds, t.c.B().Hincrby().Key(livestate.TallyKey(k.show, k.q, t.shard)).Field(opt).Increment(n).Build())
			undo = append(undo, func() { t.restore(k, opt, n) })
		}
	}
	var first error
	for i, r := range t.c.DoMulti(ctx, cmds...) {
		if err := r.Error(); err != nil {
			undo[i]()
			if first == nil {
				first = err
			}
		}
	}
	return first
}

func (t *Tally) restore(k tallyKey, opt string, n int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pending[k] == nil {
		t.pending[k] = map[string]int64{}
	}
	t.pending[k][opt] += n
}

// Sum reads every shard for (show, q) and adds them up.
func Sum(ctx context.Context, c valkey.Client, show string, q int) (map[string]int64, error) {
	cmds := make(valkey.Commands, livestate.TallyShards)
	for s := range cmds {
		cmds[s] = c.B().Hgetall().Key(livestate.TallyKey(show, q, s)).Build()
	}
	out := map[string]int64{}
	for _, r := range c.DoMulti(ctx, cmds...) {
		m, err := r.AsStrMap()
		if err != nil {
			return nil, err
		}
		for opt, v := range m {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, err
			}
			out[opt] += n
		}
	}
	return out, nil
}
