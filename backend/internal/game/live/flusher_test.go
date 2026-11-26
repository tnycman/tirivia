package live

import (
	"context"
	"testing"
	"time"
)

func TestRunFlushesUntilCancelled(t *testing.T) {
	c, show := client(t), id(t)
	ctx, cancel := context.WithCancel(context.Background())
	tl := NewTally(c, 5)
	done := make(chan struct{})
	go func() { tl.Run(ctx, 10*time.Millisecond, func(err error) { t.Error(err) }); close(done) }()
	tl.Add(show, 1, "a")
	tl.Add(show, 1, "a")
	deadline := time.Now().Add(5 * time.Second)
	for {
		got, err := Sum(context.Background(), c, show, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got["a"] == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("tally never flushed: %v", got)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
}
