package live

import (
	"context"
	"time"
)

// Run flushes every interval until ctx ends; a failed flush keeps its counts for the next tick.
func (t *Tally) Run(ctx context.Context, every time.Duration, onErr func(error)) {
	tk := time.NewTicker(every)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			if err := t.Flush(ctx); err != nil {
				onErr(err)
			}
		}
	}
}
