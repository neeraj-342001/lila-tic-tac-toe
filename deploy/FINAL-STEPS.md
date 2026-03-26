# Final deployment steps (end-to-end)

Use this after your code is on GitHub. You need: a **VM with Docker**, a **DNS name** for Nakama, and **Vercel** (or similar) for the frontend.

---

## 0. Choose values (write them down once)

| Item | You choose | Used in |
| --- | --- | --- |
| Nakama hostname | e.g. `nakama.yourdomain.com` | DNS A record, `deploy/.env`, Vercel `VITE_NAKAMA_HOST` |
| Postgres password | Long random (letters/numbers safest) | `deploy/.env` only → `POSTGRES_PASSWORD` |
| Server key | Long random string | `deploy/.env` → `SERVER_KEY` **and** Vercel → `VITE_NAKAMA_SERVER_KEY` (**must match**) |

---

## 1. DNS (before or right after the VM exists)

1. Create a **subdomain** (e.g. `nakama`) under a domain you control.
2. Add an **A record**: host `nakama` (or `@` if you use the root) → **public IPv4** of your VM.
3. Wait until it resolves (check with `ping nakama.yourdomain.com` or `dig`).

---

## 2. Server: install Docker

On Ubuntu (typical):

```bash
sudo apt update && sudo apt install -y docker.io docker-compose-plugin
sudo usermod -aG docker $USER
# log out and back in, or use newgrp docker
```

---

## 3. Server: clone repo and create `deploy/.env`

```bash
git clone https://github.com/YOUR_USERNAME/lila-tic-tac-toe.git
cd lila-tic-tac-toe/deploy
cp .env.production.example .env
nano .env
```

**Fill `deploy/.env` exactly like this (replace placeholders):**

```env
NAKAMA_PUBLIC_HOST=nakama.yourdomain.com
POSTGRES_PASSWORD=YOUR_LONG_POSTGRES_PASSWORD
SERVER_KEY=YOUR_LONG_SERVER_KEY_SAME_AS_VERCEL
```

Save the file. **Do not commit `.env`** (it is gitignored).

---

## 4. Server: open firewall

Allow **TCP 80** and **TCP 443** from the internet (Caddy + Let’s Encrypt). **Do not** expose Postgres `5432` or raw Nakama `7350` publicly.

---

## 5. Server: start the stack

```bash
cd /path/to/lila-tic-tac-toe/deploy
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
docker compose -f docker-compose.prod.yml logs -f
```

- First HTTPS fetch may take a minute while Caddy obtains a certificate for `NAKAMA_PUBLIC_HOST`.
- Check: open `https://NAKAMA_PUBLIC_HOST` in a browser — you may see Nakama’s API response or a minimal page (not the React app; that’s expected).

---

## 6. Vercel: connect repo and env

1. [vercel.com](https://vercel.com) → **Import** `lila-tic-tac-toe`.
2. **Root Directory:** `frontend`.
3. **Environment Variables** (Production):

| Name | Value |
| --- | --- |
| `VITE_NAKAMA_HOST` | **Same hostname** as `NAKAMA_PUBLIC_HOST` (e.g. `nakama.yourdomain.com`) — no `https://` |
| `VITE_NAKAMA_PORT` | `443` |
| `VITE_NAKAMA_USE_SSL` | `true` |
| `VITE_NAKAMA_SERVER_KEY` | **Exactly** the same string as `SERVER_KEY` in `deploy/.env` |

4. **Deploy** (or **Redeploy** after saving env vars).

Your **public game URL** is the Vercel URL (e.g. `https://….vercel.app`).

---

## 7. Smoke test

1. Open the Vercel URL in two browser windows (or one + incognito).
2. You should see **Online** / no auth errors in the UI.
3. Click **Find match** in both — game should start.

If the page loads but Nakama fails: recheck **host**, **443**, **SSL=true**, and **key match**. If HTTPS to Nakama fails: check DNS, firewall, and `docker compose logs caddy`.

---

## 8. What to submit (assignment)

| Deliverable | Example |
| --- | --- |
| Repository | `https://github.com/YOUR_USERNAME/lila-tic-tac-toe` |
| Live game | Your **Vercel** URL |
| Live Nakama API | `https://YOUR_NAKAMA_HOST` (same host as `NAKAMA_PUBLIC_HOST`) |

---

## Local dev vs production (reminder)

| File / place | Purpose |
| --- | --- |
| `frontend/.env.local` (from `.env.example`) | **Laptop** + local `docker compose` — `127.0.0.1`, port `7350`, `defaultkey` |
| `deploy/.env` (from `.env.production.example`) | **Server** Docker only — not for Vercel |
| Vercel env vars | **Production** frontend build — `VITE_*` |

More detail: [DEPLOYMENT.md](./DEPLOYMENT.md).
