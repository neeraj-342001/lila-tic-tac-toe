Lila Backend Assignment — submission document
==============================================

Candidate name: Neeraj Khatri  
Email: neerajkhatri180@gmail.com  
Role applied: Backend Developer  

---

IMPORTANT FOR REVIEWERS — THE VERCEL APP MAY NOT WORK “BY ITSELF”
-----------------------------------------------------------------

Please read this before opening only the live game link.

• The game at https://lila-tic-tac-toe.vercel.app/ is the **frontend** (static site). It talks to **Nakama** using the `VITE_*` environment variables set in Vercel (host, port, SSL, server key).

• In this submission there is **no always-on Nakama server in the cloud** (see “Why this architecture” below). Nakama is meant to run **on a machine via Docker**, and a **Cloudflare Tunnel** gives that machine a temporary public **https://….trycloudflare.com** URL.

• **If you open only the Vercel URL** while **no** backend is running **or** the tunnel is down **or** the tunnel hostname no longer matches what is configured in Vercel, the app will show connection / auth errors — **that is expected**, not a broken build.

**How it connects when everything is running**

1. **Docker** runs Postgres + Nakama on `http://127.0.0.1:7350` (repo root: `docker compose up --build`).  
2. **cloudflared** publishes that port to the internet as **HTTPS** at a `*.trycloudflare.com` address.  
3. **Vercel** is set so the browser calls that hostname on port **443** with **TLS** and the same **server key** as in `docker-compose.yml` (`defaultkey` for local dev).  
4. Your browser (on the Vercel page) → **internet** → **tunnel** → **local Nakama**.

**How you can test successfully**

• **Option A — Match this submission:** On a machine with Docker, clone the repo, run `docker compose up --build`, run `cloudflared tunnel --url http://127.0.0.1:7350`, copy the printed hostname, put it in Vercel env as `VITE_NAKAMA_HOST` (and keep port 443, SSL true, server key `defaultkey`), redeploy Vercel, then open the Vercel URL in two tabs.  

• **Option B — Easiest local test (no tunnel):** Same machine: run `docker compose up`, then `cd frontend && npm run dev` and use **http://localhost:5173** — both frontend and backend are local; no Vercel needed for verifying multiplayer.

**Why we did it this way**

Cloud VPS sign-up often requires **credit/debit card** verification; I did **not** use a card, so I did not host Nakama 24/7 on a cloud VM. The tunnel + local Docker approach still demonstrates a **public HTTPS Nakama endpoint** while the demo is active. Full VM deployment is documented in the repo (`deploy/FINAL-STEPS.md`) if you prefer an always-on server.

---

DELIVERABLES (PDF CHECKLIST)
-----------------------------

1) Source code repository  
   https://github.com/neeraj-342001/lila-tic-tac-toe  

2) Deployed and accessible game URL  
   https://lila-tic-tac-toe.vercel.app/  

3) Deployed Nakama server endpoint  
   See section “Nakama endpoint” below. The backend runs in Docker locally; a Cloudflare Tunnel provides a public HTTPS URL while Docker and the tunnel are running.

4) README (setup, architecture, deployment, API/config, testing)  
   Main README (same repository):  
   https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md  

   Direct links to README sections:  
   • Setup / installation — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md#quick-start-local  
   • Architecture — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md#architecture  
   • Deployment — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md#deployment-production  
   • API / server configuration — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md#api--configuration-reference  
   • How to test multiplayer — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md#how-to-test-multiplayer  

   Extra deployment docs in the repo:  
   • Production VPS path — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/deploy/FINAL-STEPS.md  
   • Local Nakama + Cloudflare Tunnel — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/deploy/LOCAL-TUNNEL.md  
   • Full deployment notes — https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/deploy/DEPLOYMENT.md  

---

WHY THIS ARCHITECTURE (SUMMARY)
-------------------------------

(See “IMPORTANT FOR REVIEWERS” above for the full picture.)

• Vercel = static frontend only.  
• Nakama + Postgres = Docker on a developer machine; Cloudflare Tunnel = temporary public HTTPS URL to that Nakama.  
• No card on file for cloud VPS → no 24/7 hosted Nakama in this submission; optional always-on path: https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/deploy/FINAL-STEPS.md  

---

NAKAMA ENDPOINT (DETAIL)
------------------------

• Run: docker compose up --build (from repo root).  
• Run in a second terminal: cloudflared tunnel --url http://127.0.0.1:7350  
• Use the printed https://____________.trycloudflare.com hostname. Update the line below when you run a new tunnel session.

Current tunnel URL (fill in when demoing):  
https://______________________________.trycloudflare.com  

The endpoint is only reachable while both Docker and cloudflared are running on the developer machine.

Vercel environment variables (must match docker-compose.yml SERVER_KEY, default “defaultkey” for local dev):  
• VITE_NAKAMA_HOST = hostname only, e.g. abc-xyz.trycloudflare.com  
• VITE_NAKAMA_PORT = 443  
• VITE_NAKAMA_USE_SSL = true  
• VITE_NAKAMA_SERVER_KEY = defaultkey  

---

QUICK LINKS (PLAIN)
-------------------

GitHub repository:  
https://github.com/neeraj-342001/lila-tic-tac-toe  

Live game (Vercel):  
https://lila-tic-tac-toe.vercel.app/  

README:  
https://github.com/neeraj-342001/lila-tic-tac-toe/blob/main/README.md  

---

OPTIONAL BONUS FEATURES (PDF)
-----------------------------

• Timer / classic vs timed mode: implemented (see backend/match.go and frontend).  
• Concurrent matches: separate Nakama match instances; no global leaderboard in this submission.  

---
End of submission document.
