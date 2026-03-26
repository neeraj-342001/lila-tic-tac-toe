# Lila — Multiplayer Tic-Tac-Toe (Nakama + React)

Take-home style project: **server-authoritative** tic-tac-toe on [Nakama](https://heroiclabs.com/docs/nakama/), with a **mobile-first React** client.

## What you get

- **Backend**: Go plugin (`backend/`) implementing an authoritative `MatchHandler`: moves are validated on the server, board state is broadcast after each valid turn, turn timers (classic vs fast), disconnect handling, match discovery via labels (`find_match` RPC + `MatchList` / `MatchCreate`).
- **Frontend**: Vite + React + TypeScript (`frontend/`) using `@heroiclabs/nakama-js` (device auth, socket, RPC, match data).
- **Local stack**: Docker Compose runs PostgreSQL + Nakama with the plugin loaded.

## Prerequisites

- Docker Desktop (or Docker Engine) for the Nakama stack
- Node.js 20+ for the web app

## Quick start (local)

1. **Start Nakama + Postgres**

   ```bash
   cd lila-tic-tac-toe
   docker compose up --build
   ```

   - Nakama API / WebSocket: `http://127.0.0.1:7350` (HTTP), `ws://127.0.0.1:7350/ws`
   - Default server key: `defaultkey` (development only)

2. **Run the React app**

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Open the printed URL (usually `http://localhost:5173`).

3. **Play a match**

   - Click **Find match** in two separate browser profiles (or one normal + one private window), or use a phone on the same LAN (set `VITE_NAKAMA_HOST` to your machine’s LAN IP).

## Environment variables (frontend)

Create `frontend/.env.local` if defaults are wrong:

```env
VITE_NAKAMA_HOST=127.0.0.1
VITE_NAKAMA_PORT=7350
VITE_NAKAMA_SERVER_KEY=defaultkey
```

For phones on the same Wi‑Fi, use your computer’s LAN IP as `VITE_NAKAMA_HOST` and ensure Nakama accepts connections on that interface (firewall).

## Architecture

| Layer | Role |
| --- | --- |
| **Nakama match loop** | Single source of truth for the 3×3 board, turn order, win/draw, timeouts, and match labels (open/closed, fast mode). |
| **RPC `find_match`** | Lists open matches with the same mode (`fast` flag) or creates a new authoritative match. |
| **WebSocket `match_data`** | Clients send opcode `4` (move) with JSON `{"position":0..8}`; server broadcasts opcodes 1–3 (start/update/done). |

Opcodes are defined in `backend/match.go` and `frontend/src/protocol.ts` and must stay in sync.

## Deployment (production)

Linear checklist (DNS → server `.env` → Vercel `VITE_*` → test): **[`deploy/FINAL-STEPS.md`](deploy/FINAL-STEPS.md)**.  
More detail: **[`deploy/DEPLOYMENT.md`](deploy/DEPLOYMENT.md)**.

**Short version**

1. **Nakama:** On a cloud VM, from the `deploy/` folder: copy `deploy/.env.production.example` → `.env`, set `NAKAMA_PUBLIC_HOST`, `POSTGRES_PASSWORD`, `SERVER_KEY`, then run  
   `docker compose -f docker-compose.prod.yml --env-file .env up -d --build`  
   (Postgres + Nakama + Caddy with HTTPS.)

2. **Frontend:** `cd frontend && npm run build`, deploy `dist/` to Vercel/Netlify/etc. Set at build time:
   - `VITE_NAKAMA_HOST` — same hostname as `NAKAMA_PUBLIC_HOST`
   - `VITE_NAKAMA_PORT=443`
   - `VITE_NAKAMA_USE_SSL=true`
   - `VITE_NAKAMA_SERVER_KEY` — **must match** `SERVER_KEY` on the server

3. Point DNS for the game host and for the Nakama host to the right targets (see `deploy/DEPLOYMENT.md`).

**Nakama notes**

- Build the Go plugin with `heroiclabs/nakama-pluginbuilder` at the **same** tag as the Nakama image (`backend/Dockerfile`).
- Use a **strong** `SERVER_KEY`; do not expose Postgres or raw Nakama ports publicly when using the provided Caddy setup.

## API / configuration reference

| Item | Local default |
| --- | --- |
| Postgres | `postgres:5432`, DB `nakama`, user `postgres`, password `localdb` |
| Nakama ports | `7350` (client API), `7351` (console) |
| Server key | `defaultkey` |

## How to test multiplayer

1. Start Docker Compose and `npm run dev`.
2. Open two browsers (or incognito + normal).
3. Both click **Find match**; when two humans are in the same open match, the server assigns **X** and **O** deterministically (sorted user IDs).
4. Try invalid moves (wrong turn, occupied cell); server rejects with opcode `5`.
5. Leave one tab: the other should see **opponent disconnected** (opcode `6`).

## Repository layout

```
lila-tic-tac-toe/
├── backend/           # Go Nakama module (plugin)
│   ├── Dockerfile
│   ├── entrypoint.sh
│   ├── main.go
│   ├── match.go
│   ├── rpc.go
│   └── local.yml
├── deploy/            # Production compose + Caddy + DEPLOYMENT.md
├── docker-compose.yml
├── frontend/          # Vite React client
└── README.md
```

## License

Apache-2.0 for structure aligned with Heroic Labs samples; your employer may require a different license for submission.
