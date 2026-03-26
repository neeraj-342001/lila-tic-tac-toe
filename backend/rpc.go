package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// findMatchRequest is sent by the client before joining a match (JSON body to RPC `find_match`).
type findMatchRequest struct {
	Fast bool `json:"fast"`
}

type findMatchResponse struct {
	MatchIDs []string `json:"match_ids"`
}

func rpcFindMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	if _, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string); !ok {
		return "", errNoUserID
	}

	var req findMatchRequest
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return "", errUnmarshal
		}
	}

	fast := 0
	if req.Fast {
		fast = 1
	}
	query := fmt.Sprintf("+label.open:1 +label.fast:%d", fast)
	maxSize := 1

	matches, err := nk.MatchList(ctx, 10, true, "", nil, &maxSize, query)
	if err != nil {
		logger.Error("match list: %v", err)
		return "", errInternalError
	}

	matchIDs := make([]string, 0, 4)
	if len(matches) > 0 {
		for _, m := range matches {
			matchIDs = append(matchIDs, m.MatchId)
		}
	} else {
		matchID, err := nk.MatchCreate(ctx, moduleName, map[string]interface{}{"fast": req.Fast})
		if err != nil {
			logger.Error("match create: %v", err)
			return "", errInternalError
		}
		matchIDs = append(matchIDs, matchID)
	}

	out, err := json.Marshal(findMatchResponse{MatchIDs: matchIDs})
	if err != nil {
		return "", errMarshal
	}
	return string(out), nil
}
