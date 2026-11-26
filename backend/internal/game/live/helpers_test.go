package live

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/platform/kv"
	"github.com/tirivia/backend/internal/platform/livestate"
)

func client(t *testing.T) valkey.Client {
	t.Helper()
	url := os.Getenv("TEST_VALKEY_URL")
	if url == "" {
		t.Skip("TEST_VALKEY_URL not set (make store-test)")
	}
	c, err := kv.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c
}

func id(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

func setPlayer(t *testing.T, c valkey.Client, show, user string, p domain.Participant) {
	t.Helper()
	life := "0"
	if p.LifeUsed {
		life = "1"
	}
	err := c.Do(context.Background(), c.B().Hset().Key(livestate.PlayerKey(show, user)).FieldValue().
		FieldValue("joined_at_q", strconv.Itoa(p.JoinedAtQ)).FieldValue("last_ok", strconv.Itoa(p.LastOK)).
		FieldValue("life_used", life).Build()).Error()
	if err != nil {
		t.Fatal(err)
	}
}

func openQuestion(q int, now time.Time) livestate.Show {
	return livestate.Show{Phase: 3, QuestionIndex: q, QuestionID: "q" + strconv.Itoa(q), TCue: now,
		Deadline: now.Add(10 * time.Second), CorrectOption: "right", CueSeq: 1}
}
