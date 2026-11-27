package livestate

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

// ErrNoShow means the show has no live state yet.
var ErrNoShow = errors.New("livestate: show not live")

// Show is the show-state hash; CorrectOption is read only by game tasks (Valkey ACLs, stack/04).
type Show struct {
	Phase         int32
	QuestionIndex int
	QuestionID    string
	TCue          time.Time
	CloseAt       time.Time
	Deadline      time.Time
	OpenCue       []byte
	CorrectOption string
	CueSeq        uint64
}

// Shows reads and writes show state.
type Shows struct{ c valkey.Client }

// NewShows returns a Shows store.
func NewShows(c valkey.Client) *Shows { return &Shows{c: c} }

// Write replaces the show state.
func (s *Shows) Write(ctx context.Context, show string, st Show) error {
	cmd := s.c.B().Hset().Key(ShowKey(show)).FieldValue().
		FieldValue("phase", strconv.Itoa(int(st.Phase))).
		FieldValue("q", strconv.Itoa(st.QuestionIndex)).
		FieldValue("qid", st.QuestionID).
		FieldValue("t_cue", strconv.FormatInt(st.TCue.UnixMilli(), 10)).
		FieldValue("close_at", strconv.FormatInt(st.CloseAt.UnixMilli(), 10)).
		FieldValue("deadline", strconv.FormatInt(st.Deadline.UnixMilli(), 10)).
		FieldValue("open", string(st.OpenCue)).
		FieldValue("correct", st.CorrectOption).
		FieldValue("seq", strconv.FormatUint(st.CueSeq, 10)).Build()
	return s.c.Do(ctx, cmd).Error()
}

// Read returns the show state.
func (s *Shows) Read(ctx context.Context, show string) (Show, error) {
	m, err := s.c.Do(ctx, s.c.B().Hgetall().Key(ShowKey(show)).Build()).AsStrMap()
	if err != nil {
		return Show{}, err
	}
	if len(m) == 0 {
		return Show{}, ErrNoShow
	}
	return parseShow(m)
}

func parseShow(m map[string]string) (Show, error) {
	phase, err1 := strconv.ParseInt(m["phase"], 10, 32)
	q, err2 := strconv.Atoi(m["q"])
	tCue, err3 := strconv.ParseInt(m["t_cue"], 10, 64)
	deadline, err4 := strconv.ParseInt(m["deadline"], 10, 64)
	seq, err5 := strconv.ParseUint(m["seq"], 10, 64)
	closeAt, err6 := strconv.ParseInt(m["close_at"], 10, 64)
	if err := errors.Join(err1, err2, err3, err4, err5, err6); err != nil {
		return Show{}, err
	}
	return Show{Phase: int32(phase), QuestionIndex: q, QuestionID: m["qid"], TCue: time.UnixMilli(tCue),
		CloseAt: time.UnixMilli(closeAt), Deadline: time.UnixMilli(deadline), OpenCue: []byte(m["open"]),
		CorrectOption: m["correct"], CueSeq: seq}, nil
}
