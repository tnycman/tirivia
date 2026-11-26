package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"github.com/valkey-io/valkey-go"

	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/contracts"
	"github.com/tirivia/backend/internal/game/live"
	"github.com/tirivia/backend/internal/platform/kv"
	"github.com/tirivia/backend/internal/platform/livestate"
	"github.com/tirivia/backend/internal/platform/switches"
)

const (
	userHeader   = "X-Test-User"
	testCueKeyID = 7
)

type headerAuthn struct{}

func (headerAuthn) Authenticate(_ context.Context, h contracts.Headers, _, _ string) (contracts.Principal, error) {
	if u := h.Get(userHeader); u != "" {
		return contracts.Principal{UserID: u}, nil
	}
	return contracts.Principal{}, contracts.ErrUnauthenticated
}

// inventory is idempotent per (user, show) like the real one must be; lostReply debits but reports an error.
type inventory struct {
	mu                          sync.Mutex
	n                           map[string]int
	consumed                    map[string]bool
	failing, lostReply, readErr bool
}

func (i *inventory) Available(_ context.Context, u string) (int, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.readErr {
		return 0, errors.New("inventory read failed")
	}
	return i.n[u], nil
}

func (i *inventory) Consume(_ context.Context, u, show string) (bool, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.consumed[u+"/"+show] {
		return true, nil
	}
	if i.failing || i.n[u] == 0 {
		return false, nil
	}
	i.n[u]--
	i.consumed[u+"/"+show] = true
	if i.lostReply {
		i.lostReply = false
		return false, errors.New("reply lost")
	}
	return true, nil
}

type game struct {
	client  gamev1connect.GameServiceClient
	kv      valkey.Client
	shows   *livestate.Shows
	players *live.Players
	inv     *inventory
	show    string
}

func newGame(t *testing.T) game {
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
	g := game{kv: c, shows: livestate.NewShows(c), players: live.NewPlayers(c, time.Hour),
		inv: &inventory{n: map[string]int{}, consumed: map[string]bool{}}, show: rnd(t)}
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	srv := connect.NewServer()
	gamev1connect.RegisterGameServiceHandler(srv, New(Deps{PublicBaseURL: ts.URL, Switches: switches.NewMemory(switches.Snapshot{}),
		Authn: headerAuthn{}, Shows: g.shows, Players: g.players, Answers: live.NewAnswers(c), Lives: live.NewLives(c),
		Inventory: g.inv, ReactionBatchInterval: 5 * time.Second,
		CueKeys: []*gamev1.CueKey{{KeyId: testCueKeyID, Ed25519PublicKey: make([]byte, 32)}}}))
	connecthttp.Mount(mux, srv)
	g.client = gamev1connect.NewGameServiceClient(connect.NewClient(connecthttp.NewTransport(ts.Client(), ts.URL)))
	return g
}

func as(user string) context.Context {
	ctx, info := connect.NewClientContext(context.Background())
	info.RequestHeader().Set(userHeader, user)
	return ctx
}

func rnd(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
