#!/bin/sh
set -e

# postgres:password@host:port/dbname?sslmode=disable
DB="${DATABASE_ADDRESS:-postgres:localdb@postgres:5432/nakama?sslmode=disable}"
KEY="${SERVER_KEY:-defaultkey}"
LOG_LEVEL="${LOG_LEVEL:-debug}"

/nakama/nakama migrate up --database.address "$DB"

exec /nakama/nakama \
  --config /nakama/data/local.yml \
  --database.address "$DB" \
  --socket.server_key "$KEY" \
  --logger.level "$LOG_LEVEL"
