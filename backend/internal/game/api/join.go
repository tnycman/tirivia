package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
	gamev1 "github.com/tirivia/protocol/gen/go/tirivia/game/v1"
	"github.com/tirivia/protocol/gen/go/tirivia/game/v1/gamev1connect"

	"github.com/tirivia/backend/internal/platform/livestate"
	"github.com/tirivia/backend/internal/platform/switches"
)

var errNotLive = connect.NewError(connect.CodeNotFound, "show is not live")

// joinable is the join window, LOBBY through FINAL (UF-16).
var joinable = map[int32]bool{
	int32(cuev1.Phase_PHASE_LOBBY): true, int32(cuev1.Phase_PHASE_QUESTION_OPEN): true, int32(cuev1.Phase_PHASE_QUESTION_CLOSED): true,
	int32(cuev1.Phase_PHASE_REVEAL): true, int32(cuev1.Phase_PHASE_FINAL): true,
}

// JoinShow joins at the question the server says is open (AC-8, UF-16); playback_url, reaction_set and eligibility come in later slices.
func (h *Handler) JoinShow(ctx context.Context, req *gamev1.JoinShowRequest) (*gamev1.JoinShowResponse, error) {
	p, err := h.caller(ctx, gamev1connect.GameServiceJoinShowProcedure)
	if err != nil {
		return nil, err
	}
	if req.GetShowId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, "show_id is required")
	}
	st, err := h.d.Shows.Read(ctx, req.GetShowId())
	if errors.Is(err, livestate.ErrNoShow) || (err == nil && !joinable[st.Phase]) {
		return nil, errNotLive
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	opened := st.QuestionIndex
	if st.Phase == int32(cuev1.Phase_PHASE_LOBBY) {
		opened = 0
	}
	_, seed, err := h.d.Players.JoinSeeded(ctx, req.GetShowId(), p.UserID, opened)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	ps, err := h.playerState(ctx, req.GetShowId(), p.UserID, st)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "retry")
	}
	chat := h.d.Switches.Enabled(ctx, switches.ChatEnabled, req.GetShowId())
	out := &gamev1.JoinShowResponse{PlayerState: ps, ShuffleSeed: seed, CueKeys: h.d.CueKeys, ChatEnabled: chat,
		ReactionBatchIntervalMs: uint32(h.d.ReactionBatchInterval.Milliseconds())} //nolint:gosec // G115: a few seconds
	if chat && h.d.Chat != nil { // a failed mint leaves the token empty; the app then calls GetChatToken
		out.ChatToken, out.ChatEndpoint, _ = h.d.Chat.MintChatToken(ctx, req.GetShowId())
	}
	return out, nil
}
