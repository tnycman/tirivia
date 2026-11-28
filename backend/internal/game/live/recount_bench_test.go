package live

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/platform/kv"
	"trivia.al/backend/internal/platform/livestate"
)

// loadTarget is the TS-15 load-test player count.
const loadTarget = 600_000

// BenchmarkRecount measures the post-reveal recount at the load-test target; record results in research 09.
func BenchmarkRecount(b *testing.B) {
	url := os.Getenv("TEST_VALKEY_URL")
	if url == "" {
		b.Skip("TEST_VALKEY_URL not set (make store-test)")
	}
	c, err := kv.Open(url)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(c.Close)
	ctx, show := context.Background(), "bench-"+strconv.Itoa(b.N)
	cmds := make(valkey.Commands, 0, 1000)
	for i := range loadTarget {
		user := "u" + strconv.Itoa(i)
		key := livestate.CountedKey(show, 1, livestate.Bucket(user))
		cmds = append(cmds, c.B().Hset().Key(key).FieldValue().FieldValue(user, "right").Build())
		if len(cmds) == cap(cmds) || i == loadTarget-1 {
			for _, r := range c.DoMulti(ctx, cmds...) {
				if err := r.Error(); err != nil {
					b.Fatal(err)
				}
			}
			cmds = cmds[:0]
		}
	}
	b.ResetTimer()
	for b.Loop() {
		got, err := Recount(ctx, c, show, 1)
		if err != nil || got["right"] != loadTarget {
			b.Fatalf("recount = %v, %v", got, err)
		}
	}
	b.StopTimer()
	for bucket := range livestate.TallyShards {
		_ = c.Do(ctx, c.B().Del().Key(livestate.CountedKey(show, 1, bucket)).Build()).Error()
	}
}
