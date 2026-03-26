# Deploying Lila Tic-Tac-Toe (assignment checklist)

This covers a **public Nakama endpoint** and a **public static frontend**, as required by the take-home PDF.

## Architecture (recommended)

| Piece | Where | Notes |
| --- | --- | --- |
| **Postgres** | Same VM as Nakama (Docker network only, not exposed) | Managed DB (RDS, etc.) is fine too — set `DATABASE_ADDRESS` accordingly. |
| **Nakama + Go plugin** | VPS (e.g. DigitalOcean, AWS EC2, GCP VM) | Built from `backend/Dockerfile`. |
| **HTTPS** | [Caddy](https://caddyserver.com/) in `docker-compose.prod.yml` | Terminates TLS; proxies to Nakama `7350`. |
| **Frontend** | Static host (Vercel, Netlify, S3+CloudFront, GitHub Pages) | `npm run build` → upload `frontend/dist/`. |

Browsers require **HTTPS** for the game page when talking to a **secure** API; use `wss://` for the socket. The React app uses `VITE_NAKAMA_USE_SSL=true` with host/port `443` (or your TLS port).

---

## 1. Deploy Nakama (VPS + Docker)

**Requirements:** Ubuntu 22.04+ (or similar), Docker + Docker Compose v2, DNS pointed at the VM.

1. Copy the repo to the server (git clone or `scp`).

2. Configure secrets:

   ```bash
   cd deploy
   cp .env.production.example .env
   ```

   Edit `.env`:

   - `NAKAMA_PUBLIC_HOST` — hostname for Nakama (e.g. `nakama.yourdomain.com`). Create an **A record** to this server’s public IP.
   - `POSTGRES_PASSWORD` — long random string (avoid `&`, `#`, `@` in passwords or URL-encode when embedding in `DATABASE_ADDRESS`).
   - `SERVER_KEY` — long random secret; **must match** the frontend build (`VITE_NAKAMA_SERVER_KEY`).

3. Start the stack:

   ```bash
   docker compose -f docker-compose.prod.yml --env-file .env up -d --build
   ```

4. Check logs:

   ```bash
   docker compose -f docker-compose.prod.yml logs -f nakama caddy
   ```

   Caddy will obtain a Let’s Encrypt certificate when port **80** and **443** are reachable on the public IP and DNS matches `NAKAMA_PUBLIC_HOST`.

5. **Firewall:** allow **80/tcp** and **443/tcp** from the internet. Do **not** expose Postgres (`5432`) or Nakama **7350** directly; only Caddy should be public.

**Nakama console (7351)** is not published in this compose file. For emergencies, use `docker compose exec` or temporarily add a port mapping (bind to `127.0.0.1` only) and SSH tunnel.

---

## 2. Build and deploy the frontend

At build time, Vite bakes in `VITE_*` variables.

Example for **Vercel** (or Netlify / manual `npm run build`):

| Variable | Example (production) |
| --- | --- |
| `VITE_NAKAMA_HOST` | `nakama.yourdomain.com` |
| `VITE_NAKAMA_PORT` | `443` |
| `VITE_NAKAMA_USE_SSL` | `true` |
| `VITE_NAKAMA_SERVER_KEY` | same value as `SERVER_KEY` in `deploy/.env` |

Commands:

```bash
cd frontend
npm ci
npm run build
```

Upload `frontend/dist/` to your static host, or connect the repo to Vercel/Netlify and set the env vars in the dashboard.

---

## 3. CORS / mixed content

If the browser blocks requests from `https://your-game.vercel.app` to `https://nakama.yourdomain.com`, check:

- Frontend uses **https** + **wss** (`VITE_NAKAMA_USE_SSL=true`).
- No mixed content (HTTP API from HTTPS page).

Nakama often works with browser clients when the API is reachable; if you still see CORS errors, add headers in **Caddy** (e.g. `header Access-Control-Allow-Origin`) or consult [Nakama configuration](https://heroiclabs.com/docs/nakama/getting-started/configuration/).

---

## 4. Deliverables (PDF)

- **Repo:** push to GitHub/GitLab.
- **Live game URL:** your static site URL.
- **Live Nakama URL:** `https://<NAKAMA_PUBLIC_HOST>` (API + WebSocket on standard HTTPS port).
- **README:** root `README.md` + this file document setup and deployment.

---

## 5. Alternatives

- **Managed Postgres:** set `DATABASE_ADDRESS` in a custom `docker-compose` override to your cloud DSN (with `sslmode=require` if required).
- **Heroic Cloud / other hosted Nakama:** build `backend.so` with the **same Nakama version** as the server, upload the module per provider docs.
- **Official reference:** [Deploy on DigitalOcean](https://heroiclabs.com/docs/nakama/guides/deployment/digital-ocean/) (Heroic Labs).
