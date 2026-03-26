package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"math/rand"
	"sort"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	moduleName = "tic_tac_toe"

	tickRate = 5

	maxEmptySec = 120

	delayBetweenGamesSec = 5
	turnTimeFastSec      = 10
	turnTimeNormalSec    = 20
)

// Opcodes (JSON payloads; binary opcode is int64 on the wire).
const (
	opStart         = 1
	opUpdate        = 2
	opDone          = 3
	opMove          = 4
	opRejected      = 5
	opOpponentLeft  = 6
)

var winningPositions = [][]int32{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
	{0, 4, 8}, {2, 4, 6},
}

var _ runtime.Match = (*MatchHandler)(nil)

type MatchHandler struct{}

// MatchLabel is indexed by Nakama for match listing / discovery queries.
type MatchLabel struct {
	Open int `json:"open"`
	Fast int `json:"fast"`
}

// MatchState is the authoritative game state (never trust the client).
type MatchState struct {
	random *rand.Rand
	label  *MatchLabel

	emptyTicks int

	presences       map[string]runtime.Presence
	joinsInProgress int

	playing bool
	board   []int            // 0 empty, 1 X, 2 O
	marks   map[string]int   // userID -> 1 (X) or 2 (O)
	mark    int              // whose turn: 1 or 2
	deadlineRemainingTicks int64
	winner                 int
	winnerPositions        []int32
	nextGameRemainingTicks int64
}

func (s *MatchState) connectedCount() int {
	n := 0
	for _, p := range s.presences {
		if p != nil {
			n++
		}
	}
	return n
}

func (m *MatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	fast := false
	if v, ok := params["fast"].(bool); ok {
		fast = v
	}

	label := &MatchLabel{Open: 1, Fast: 0}
	if fast {
		label.Fast = 1
	}
	labelJSON, err := json.Marshal(label)
	if err != nil {
		logger.Error("match label: %v", err)
		labelJSON = []byte(`{"open":1,"fast":0}`)
	}

	st := &MatchState{
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
		label:     label,
		presences: make(map[string]runtime.Presence, 2),
	}
	return st, tickRate, string(labelJSON)
}

func (m *MatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	s := state.(*MatchState)

	if existing, ok := s.presences[presence.GetUserId()]; ok {
		if existing == nil {
			s.joinsInProgress++
			return s, true, ""
		}
		return s, false, "already joined"
	}

	if s.connectedCount()+s.joinsInProgress >= 2 {
		return s, false, "match full"
	}

	s.joinsInProgress++
	return s, true, ""
}

func (m *MatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*MatchState)
	t := time.Now().UTC()

	for _, p := range presences {
		s.emptyTicks = 0
		s.presences[p.GetUserId()] = p
		s.joinsInProgress--

		if s.playing {
			msg, _ := json.Marshal(map[string]interface{}{
				"board":    s.board,
				"mark":     s.mark,
				"deadline": t.Add(time.Duration(s.deadlineRemainingTicks/tickRate) * time.Second).Unix(),
			})
			_ = dispatcher.BroadcastMessage(int64(opUpdate), msg, []runtime.Presence{p}, nil, true)
		} else if s.board != nil && s.marks != nil && s.marks[p.GetUserId()] != 0 {
			msg, _ := json.Marshal(map[string]interface{}{
				"board":           s.board,
				"winner":          s.winner,
				"winner_positions": s.winnerPositions,
				"next_game_start":  t.Add(time.Duration(s.nextGameRemainingTicks/tickRate) * time.Second).Unix(),
			})
			_ = dispatcher.BroadcastMessage(int64(opDone), msg, []runtime.Presence{p}, nil, true)
		}
	}

	if s.connectedCount() >= 2 && s.label.Open != 0 {
		s.label.Open = 0
		if b, err := json.Marshal(s.label); err == nil {
			_ = dispatcher.MatchLabelUpdate(string(b))
		}
	}

	return s
}

func (m *MatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	s := state.(*MatchState)
	for _, p := range presences {
		s.presences[p.GetUserId()] = nil
	}

	var remaining []runtime.Presence
	for _, pr := range s.presences {
		if pr != nil {
			remaining = append(remaining, pr)
		}
	}
	if len(remaining) == 1 {
		_ = dispatcher.BroadcastMessage(int64(opOpponentLeft), nil, remaining, nil, true)
	}

	return s
}

func (m *MatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	s := state.(*MatchState)

	if s.connectedCount()+s.joinsInProgress == 0 {
		s.emptyTicks++
		if s.emptyTicks >= maxEmptySec*tickRate {
			logger.Info("closing idle match")
			return nil
		}
	}

	t := time.Now().UTC()

	if !s.playing {
		for uid, p := range s.presences {
			if p == nil {
				delete(s.presences, uid)
			}
		}

		if s.connectedCount() < 2 && s.label.Open != 1 {
			s.label.Open = 1
			if b, err := json.Marshal(s.label); err == nil {
				_ = dispatcher.MatchLabelUpdate(string(b))
			}
		}

		if s.connectedCount() < 2 {
			return s
		}

		if s.nextGameRemainingTicks > 0 {
			s.nextGameRemainingTicks--
			return s
		}

		s.playing = true
		s.board = make([]int, 9)
		s.marks = make(map[string]int, 2)
		uids := make([]string, 0, 2)
		for uid, p := range s.presences {
			if p != nil {
				uids = append(uids, uid)
			}
		}
		sort.Strings(uids)
		if len(uids) >= 2 {
			s.marks[uids[0]] = 1
			s.marks[uids[1]] = 2
		}
		s.mark = 1
		s.winner = 0
		s.winnerPositions = nil
		s.deadlineRemainingTicks = deadlineTicks(s.label)
		s.nextGameRemainingTicks = 0

		startPayload, _ := json.Marshal(map[string]interface{}{
			"board":    s.board,
			"marks":    s.marks,
			"mark":     s.mark,
			"deadline": t.Add(time.Duration(s.deadlineRemainingTicks/tickRate) * time.Second).Unix(),
		})
		_ = dispatcher.BroadcastMessage(int64(opStart), startPayload, nil, nil, true)
		return s
	}

	for _, message := range messages {
		if message.GetOpCode() != int64(opMove) {
			_ = dispatcher.BroadcastMessage(int64(opRejected), nil, []runtime.Presence{message}, nil, true)
			continue
		}

		uid := message.GetUserId()
		playerMark := s.marks[uid]
		if playerMark == 0 || s.mark != playerMark {
			_ = dispatcher.BroadcastMessage(int64(opRejected), nil, []runtime.Presence{message}, nil, true)
			continue
		}

		var body struct {
			Position int `json:"position"`
		}
		if err := json.Unmarshal(message.GetData(), &body); err != nil {
			_ = dispatcher.BroadcastMessage(int64(opRejected), nil, []runtime.Presence{message}, nil, true)
			continue
		}
		pos := body.Position
		if pos < 0 || pos > 8 || s.board[pos] != 0 {
			_ = dispatcher.BroadcastMessage(int64(opRejected), nil, []runtime.Presence{message}, nil, true)
			continue
		}

		s.board[pos] = playerMark
		switch playerMark {
		case 1:
			s.mark = 2
		case 2:
			s.mark = 1
		}
		s.deadlineRemainingTicks = deadlineTicks(s.label)

		winMark := checkWinner(s.board)
		tie := boardFull(s.board)

		if winMark != 0 {
			s.winner = winMark
			s.winnerPositions = winningLine(s.board, winMark)
			s.playing = false
			s.deadlineRemainingTicks = 0
			s.nextGameRemainingTicks = delayBetweenGamesSec * tickRate
			donePayload, _ := json.Marshal(map[string]interface{}{
				"board":            s.board,
				"winner":           s.winner,
				"winner_positions": s.winnerPositions,
				"next_game_start":  t.Add(time.Duration(s.nextGameRemainingTicks/tickRate) * time.Second).Unix(),
			})
			_ = dispatcher.BroadcastMessage(int64(opDone), donePayload, nil, nil, true)
			continue
		}
		if tie {
			s.winner = 0
			s.winnerPositions = nil
			s.playing = false
			s.deadlineRemainingTicks = 0
			s.nextGameRemainingTicks = delayBetweenGamesSec * tickRate
			donePayload, _ := json.Marshal(map[string]interface{}{
				"board":           s.board,
				"winner":          0,
				"next_game_start": t.Add(time.Duration(s.nextGameRemainingTicks/tickRate) * time.Second).Unix(),
			})
			_ = dispatcher.BroadcastMessage(int64(opDone), donePayload, nil, nil, true)
			continue
		}

		upd, _ := json.Marshal(map[string]interface{}{
			"board":    s.board,
			"mark":     s.mark,
			"deadline": t.Add(time.Duration(s.deadlineRemainingTicks/tickRate) * time.Second).Unix(),
		})
		_ = dispatcher.BroadcastMessage(int64(opUpdate), upd, nil, nil, true)
	}

	if s.playing {
		s.deadlineRemainingTicks--
		if s.deadlineRemainingTicks <= 0 {
			s.playing = false
			switch s.mark {
			case 1:
				s.winner = 2
			case 2:
				s.winner = 1
			default:
				s.winner = 0
			}
			s.deadlineRemainingTicks = 0
			s.nextGameRemainingTicks = delayBetweenGamesSec * tickRate
			donePayload, _ := json.Marshal(map[string]interface{}{
				"board":           s.board,
				"winner":          s.winner,
				"timeout":         true,
				"next_game_start": t.Add(time.Duration(s.nextGameRemainingTicks/tickRate) * time.Second).Unix(),
			})
			_ = dispatcher.BroadcastMessage(int64(opDone), donePayload, nil, nil, true)
		}
	}

	return s
}

func (m *MatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

func (m *MatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return state
}

func deadlineTicks(l *MatchLabel) int64 {
	if l.Fast == 1 {
		return int64(turnTimeFastSec * tickRate)
	}
	return int64(turnTimeNormalSec * tickRate)
}

func checkWinner(b []int) int {
	for _, line := range winningPositions {
		m := b[line[0]]
		if m == 0 {
			continue
		}
		if b[line[1]] == m && b[line[2]] == m {
			return int(m)
		}
	}
	return 0
}

func boardFull(b []int) bool {
	for _, v := range b {
		if v == 0 {
			return false
		}
	}
	return true
}

func winningLine(b []int, mark int) []int32 {
	for _, line := range winningPositions {
		if b[line[0]] == mark && b[line[1]] == mark && b[line[2]] == mark {
			return line
		}
	}
	return nil
}
