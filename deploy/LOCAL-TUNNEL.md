# Expose local Nakama with Cloudflare Tunnel (no cloud VPS / no card)

Use this when you run **Postgres + Nakama in Docker on your machine** and want a **public HTTPS URL** so the **Vercel-hosted frontend** can reach your backend.

**Tradeoffs**

- Works **only while** Docker + `cloudflared` are running on your computer.
- **Quick Tunnel** URLs change each time you restart `cloudflared` (unless you configure a named tunnel + custom domain in Cloudflare).
- Fine for **demos / review windows**; not the same as a 24/7 datacenter deployment.

---

## 1. Start the local stack

From the **repo root** (same as normal local dev):

```bash
cd /path/to/lila-tic-tac-toe
docker compose up --build
```

Wait until Nakama logs show **Startup done**. API is at `http://127.0.0.1:7350` with server key **`defaultkey`** (see `docker-compose.yml`).

---

## 2. Install `cloudflared` (macOS)

```bash
brew install cloudflared
```

([Other platforms](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/install-and-setup/installation/))

---

## 3. Start a Quick Tunnel

In a **second** terminal:

```bash
cloudflared tunnel --url http://127.0.0.1:7350
```

Copy the printed URL, e.g. `https://random-name.trycloudflare.com`

- **Host only** (for Vercel): `random-name.trycloudflare.com`  
- **Port:** `443`  
- **SSL:** `true`  
- **Server key:** `defaultkey` (must match local `SERVER_KEY` in `docker-compose.yml`)

---

## 4. Point Vercel at the tunnel

In **Vercel → Project → Settings → Environment Variables** (Production):

| Name | Value |
| --- | --- |
| `VITE_NAKAMA_HOST` | `random-name.trycloudflare.com` (your tunnel host, no `https://`) |
| `VITE_NAKAMA_PORT` | `443` |
| `VITE_NAKAMA_USE_SSL` | `true` |
| `VITE_NAKAMA_SERVER_KEY` | `defaultkey` |

**Redeploy** the project after saving.

---

## 5. Test

1. Keep **Docker Compose** and **`cloudflared`** running.
2. Open your **Vercel URL** in two tabs → **Find match**.

---

## If something fails

- **Mixed content / connection errors:** Ensure `VITE_NAKAMA_USE_SSL=true` and port `443` when using `https://` tunnels.
- **Tunnel URL changed:** Re-copy host from `cloudflared` output, update Vercel env, redeploy.
- **Stronger server key:** Change `SERVER_KEY` in `docker-compose.yml` and match `VITE_NAKAMA_SERVER_KEY` (then rebuild containers).

---

## Security note

`defaultkey` is for **development only**. For anything beyond a short demo, use a long random `SERVER_KEY` in compose and the same value in Vercel.
