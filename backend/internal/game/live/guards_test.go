package live

import (
	"context"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// TestMalformedKeysWriteNothing: a script that meets a bad tally or counted key fails before its first write.
func TestMalformedKeysWriteNothing(t *testing.T) {
	c := client(t)
	ctx, now := context.Background(), time.Now()
	cases := map[string]func(show, user string) error{
		"tally is a string": func(show, user string) error {
			return c.Do(ctx, c.B().Set().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).Value("x").Build()).Error()
		},
		"tally field not an integer": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "x").Build()).Error()
		},
		"counted is a list": func(show, user string) error {
			return c.Do(ctx, c.B().Rpush().Key(livestate.CountedKey(show, 1, livestate.Bucket(user))).Element("x").Build()).Error()
		},
	}
	for name, corrupt := range cases {
		show, user := id(t), id(t)
		setPlayer(t, c, show, user, domain.Participant{})
		if err := corrupt(show, user); err != nil {
			t.Fatal(err)
		}
		if _, err := NewAnswers(c).Submit(ctx, show, user, openQuestion(1, now), "q1", "right", false, now); err == nil {
			t.Fatalf("%s: intake accepted", name)
		}
		if n, _ := c.Do(ctx, c.B().Hexists().Key(livestate.PlayerKey(show, user)).Field(livestate.AnswerField(1)).Build()).AsInt64(); n != 0 {
			t.Fatalf("%s: the answer was written before the failure, so a retry would never count it", name)
		}
	}
}

// TestMalformedTallyBlocksLife: life use and refund leave the player untouched when the tally is malformed.
func TestMalformedTallyBlocksLife(t *testing.T) {
	c := client(t)
	ctx, show, user := context.Background(), id(t), id(t)
	lives, players := NewLives(c), NewPlayers(c, time.Hour)
	setPlayer(t, c, show, user, domain.Participant{LastOK: 0})
	key := livestate.TallyKey(show, 1, livestate.Bucket(user))
	if err := c.Do(ctx, c.B().Hset().Key(key).FieldValue().FieldValue(LifeOption, "x").Build()).Error(); err != nil {
		t.Fatal(err)
	}
	if err := lives.Use(ctx, show, user, 1, "q1"); err == nil {
		t.Fatal("life accepted on a malformed tally")
	}
	if got, _ := players.Get(ctx, show, user); got.LifeUsed || got.LastOK != 0 {
		t.Fatalf("player changed before the failure: %+v", got)
	}
}
