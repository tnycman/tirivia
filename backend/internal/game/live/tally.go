package live

import (
	"context"
	"strconv"

	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/platform/livestate"
)

// LifeOption is the tally field counting extra lives used; players remaining = correct + lives (SPEC §6.3).
const LifeOption = "_life"

// Sum reads every bucket's tally for (show, q) and adds them up.
func Sum(ctx context.Context, c valkey.Client, show string, q int) (map[string]int64, error) {
	cmds := make(valkey.Commands, livestate.TallyShards)
	for b := range cmds {
		cmds[b] = c.B().Hgetall().Key(livestate.TallyKey(show, q, b)).Build()
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
			if n != 0 {
				out[opt] += n
			}
		}
	}
	return out, nil
}
