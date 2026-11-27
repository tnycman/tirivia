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
		"tally field is fractional": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "1.5").Build()).Error()
		},
		"tally field is exponent notation": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "1e3").Build()).Error()
		},
		"tally field is negative zero": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "-0").Build()).Error()
		},
		"tally field has a leading zero": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "007").Build()).Error()
		},
		"tally field would overflow": func(show, user string) error {
			return c.Do(ctx, c.B().Hset().Key(livestate.TallyKey(show, 1, livestate.Bucket(user))).FieldValue().FieldValue("right", "9223372036854775807").Build()).Error()
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
	ctx := context.Background()
	for name, value := range map[string]string{"not integer": "x", "would overflow": "9223372036854775807"} {
		show, user := id(t), id(t)
		lives, players := NewLives(c), NewPlayers(c, time.Hour)
		setPlayer(t, c, show, user, domain.Participant{LastOK: 0})
		key := livestate.TallyKey(show, 1, livestate.Bucket(user))
		if err := c.Do(ctx, c.B().Hset().Key(key).FieldValue().FieldValue(LifeOption, value).Build()).Error(); err != nil {
			t.Fatal(err)
		}
		if err := lives.Use(ctx, show, user, 1, "q1"); err == nil {
			t.Fatalf("%s: life accepted on a malformed tally", name)
		}
		if got, _ := players.Get(ctx, show, user); got.LifeUsed || got.LastOK != 0 {
			t.Fatalf("%s: player changed before the failure: %+v", name, got)
		}
		if got, err := c.Do(ctx, c.B().Hget().Key(key).Field(LifeOption).Build()).ToString(); err != nil || got != value {
			t.Fatalf("%s: tally changed before the failure: %q %v", name, got, err)
		}
	}
}

func TestMalformedTallyBlocksLifeRefund(t *testing.T) {
	c := client(t)
	ctx, show, user := context.Background(), id(t), id(t)
	lives, players := NewLives(c), NewPlayers(c, time.Hour)
	setPlayer(t, c, show, user, domain.Participant{LastOK: 1, LifeUsed: true})
	playerKey := livestate.PlayerKey(show, user)
	if err := c.Do(ctx, c.B().Hset().Key(playerKey).FieldValue().FieldValue("life_q", "1").FieldValue("life_qid", "q1").Build()).Error(); err != nil {
		t.Fatal(err)
	}
	tallyKey := livestate.TallyKey(show, 1, livestate.Bucket(user))
	if err := c.Do(ctx, c.B().Hset().Key(tallyKey).FieldValue().FieldValue(LifeOption, "-9223372036854775808").Build()).Error(); err != nil {
		t.Fatal(err)
	}
	if err := lives.Refund(ctx, show, user, 1); err == nil {
		t.Fatal("life refunded on a tally that would underflow")
	}
	if got, _ := players.Get(ctx, show, user); !got.LifeUsed || got.LastOK != 1 {
		t.Fatalf("player changed before the failure: %+v", got)
	}
	if got, err := c.Do(ctx, c.B().Hget().Key(tallyKey).Field(LifeOption).Build()).ToString(); err != nil || got != "-9223372036854775808" {
		t.Fatalf("tally changed before the failure: %q %v", got, err)
	}
	if q, qid, spent, err := lives.Spent(ctx, show, user); err != nil || !spent || q != 1 || qid != "q1" {
		t.Fatalf("life markers changed before the failure: %d %q %v %v", q, qid, spent, err)
	}
}
