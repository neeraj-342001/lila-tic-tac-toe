// Lila Tic-Tac-Toe — Nakama Go runtime entrypoint (server-authoritative match + RPCs).
package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	start := time.Now()

	if err := initializer.RegisterRpc("find_match", rpcFindMatch); err != nil {
		return err
	}

	if err := initializer.RegisterMatch(moduleName, func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &MatchHandler{}, nil
	}); err != nil {
		return err
	}

	logger.Info("Lila tic-tac-toe module loaded in %d ms", time.Since(start).Milliseconds())
	return nil
}
