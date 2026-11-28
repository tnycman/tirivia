package api

import (
	"context"
	"errors"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	cuev1 "trivia.al/protocol/gen/go/trivia/cue/v1"
	gamev1 "trivia.al/protocol/gen/go/trivia/game/v1"
	"trivia.al/protocol/gen/go/trivia/game/v1/gamev1connect"

	"trivia.al/backend/internal/game/domain"
	"trivia.al/backend/internal/game/live"
	"trivia.al/backend/internal/platform/livestate"
)

var statuses = map[domain.Status]gamev1.PlayerStatus{
	domain.StatusAlive:      gamev1.PlayerStatus_PLAYER_STATUS_ALIVE,
	domain.StatusEliminated: gamev1.PlayerStatus_PLAYER_STATUS_ELIMINATED,
	domain.StatusPlayAlong:  gamev1.PlayerStatus_PLAYER_STATUS_PLAY_ALONG,
}

// GetShowState lets a client resynchronise after a cue gap or a late join (SPEC §6.4).
func (h *Handler) GetShowState(ctx context.Context, req *gamev1.GetShowStateRequest) (*gamev1.GetShowStateResponse, error) {
	p, err := h.caller(ctx, gamev1connect.GameServiceGetShowStateProcedure)
	if err != nil {
		return nil, err
	}
	st, err := h.d.Shows.Read(ctx, req.GetShowId())
	if errors.Is(err, livestate.ErrNoShow) {
		return nil, connect.NewError(connect.CodeNotFound, "show is not live")
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "state unavailable; retry")
	}
	ps, err := h.playerState(ctx, req.GetShowId(), p.UserID, st)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, "state unavailable; retry")
	}
	out := &gamev1.GetShowStateResponse{Phase: cuev1.Phase(st.Phase), LastCueSeq: st.CueSeq, PlayerState: ps}
	if st.Phase == int32(cuev1.Phase_PHASE_QUESTION_OPEN) && len(st.OpenCue) > 0 {
		q := &cuev1.QuestionOpen{}
		if err := proto.Unmarshal(st.OpenCue, q); err != nil {
			return nil, connect.NewError(connect.CodeInternal, "open question unreadable")
		}
		out.OpenQuestion = q
		if left := st.CloseAt.Sub(h.d.Now()).Milliseconds(); left > 0 {
			out.RemainingMs = uint32(left) //nolint:gosec // G115: bounded by the 10 s question timer
		}
	}
	return out, nil
}

func (h *Handler) playerState(ctx context.Context, show, user string, st livestate.Show) (*gamev1.PlayerState, error) {
	pt, err := h.d.Players.Get(ctx, show, user)
	if errors.Is(err, live.ErrNotJoined) {
		return &gamev1.PlayerState{Status: gamev1.PlayerStatus_PLAYER_STATUS_SPECTATOR}, nil
	}
	if err != nil {
		return nil, err
	}
	revealed := st.Phase >= int32(cuev1.Phase_PHASE_REVEAL)
	ps := &gamev1.PlayerState{Status: statuses[pt.StatusAt(st.QuestionIndex, revealed)], ExtraLifeUsedThisShow: pt.LifeUsed}
	if h.d.Inventory != nil {
		n, err := h.d.Inventory.Available(ctx, user)
		if err != nil {
			return nil, err
		}
		ps.ExtraLivesAvailable = uint32(max(n, 0)) //nolint:gosec // G115: clamped at zero
	}
	return ps, nil
}
