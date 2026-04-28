# Deploying ChessMaster Pro to Railway

This guide spins up three Railway services from one repo:

1. **postgres** — managed Postgres
2. **chessmaster-api** — Go backend with bundled Stockfish (`Dockerfile`)
3. **chessmaster-web** — Next.js frontend (`web/Dockerfile`)

The API auto-runs schema migrations on first start (via the embedded
`internal/adapters/postgres/migrations.sql`). No `migrate` CLI step.

## Prerequisites

- A Railway account (https://railway.com)
- This repo pushed to GitHub
- A Google AI Studio API key for Gemini (free tier is fine)

## 1. Create the project

```
Railway → New Project → Empty Project → name: chessmaster
```

## 2. Add Postgres

```
+ New → Database → Add PostgreSQL
```

Railway will provision it and expose `DATABASE_URL` in that service's
**Variables** tab. Copy that value — you'll paste it into the API service.

## 3. Deploy the API

```
+ New → GitHub Repo → rwrrioe/chessmaster
```

Settings to confirm:

- **Root Directory** — leave blank (the API uses the repo root `Dockerfile`)
- **Builder** — Railway auto-detects Dockerfile via `railway.json`
- **Healthcheck** — `/healthz` (already in `railway.json`)

Then **Variables** → add:

| Name | Value |
|---|---|
| `POSTGRES_URL` | the Postgres `DATABASE_URL`, append `?sslmode=require` if not present |
| `JWT_SECRET` | output of `openssl rand -hex 32` |
| `GEMINI_API_KEY` | from https://aistudio.google.com/app/apikey |
| `GEMINI_MODEL` | `gemini-2.0-flash` (optional override) |
| `STOCKFISH_PATH` | `/usr/games/stockfish` |
| `CORS_ORIGINS` | leave blank for now; you'll set after the web service has a domain |

`PORT` is injected by Railway automatically — do not set it.

**Settings → Networking → Generate Domain** → you'll get
`https://chessmaster-api-xxxx.up.railway.app`. Verify:

```
curl https://chessmaster-api-xxxx.up.railway.app/healthz
# → {"status":"ok"}
```

Tail logs to confirm `using postgres repos` and `stockfish engine at /usr/games/stockfish`.

## 4. Deploy the frontend

```
+ New → GitHub Repo → rwrrioe/chessmaster
```

Settings to change:

- **Root Directory**: `web`
- **Build args** (Settings → Build → Build args) — these bake into the
  Next.js bundle at build time, not runtime:

| Name | Value |
|---|---|
| `NEXT_PUBLIC_API_URL` | the API domain from step 3 (https) |
| `NEXT_PUBLIC_WS_URL` | same domain with `wss://` scheme |

(Railway also exposes them as runtime env vars from the same Variables
tab — that's fine, harmless.)

**Settings → Networking → Generate Domain** → you'll get
`https://chessmaster-web-xxxx.up.railway.app`.

## 5. Lock down CORS

Go back to the **chessmaster-api** service → **Variables** → set:

```
CORS_ORIGINS=https://chessmaster-web-xxxx.up.railway.app
```

Save. Railway redeploys the api. The frontend now talks to the API and
nothing else can.

## 6. Smoke test

Open the web domain and:

1. Register an account → nav switches to **Profile / Sign out**
2. **/play** → AI Easy → drag e2 → e4 → AI replies in ~200 ms
3. **/leaderboard** → empty list (no crash)
4. End a game → **AI Coach analysis** returns Gemini response

If WS in PvP fails, check the browser console — most common cause is a
missing `wss://` scheme in `NEXT_PUBLIC_WS_URL`.

## Costs

Free tier covers all three services for hobby use:

- Postgres: 500 MB storage, 5 GB egress
- API: ~$0–5/month under light traffic
- Web: ~$0–3/month

Railway sleeps services after inactivity on the free plan; first request
spins them back up in ~5 s.

## Operations

### Logs
```
railway logs --service chessmaster-api
```

### Connect to Postgres
```
railway connect postgres
\dt          -- list tables
SELECT * FROM players LIMIT 5;
```

### Redeploy after a git push
Railway redeploys automatically on push to `main`. Watch the deploy log
for `gemini coach enabled` and `stockfish engine at /usr/games/stockfish`.

### Roll back
**Deployments** tab → pick a previous successful build → **Redeploy**.

## Troubleshooting

**"stockfish not found; AI moves disabled" in logs**
The Dockerfile installs stockfish via apt. If you switched to an alpine
base, the package isn't available — keep the Debian base.

**HTTP 503 from /games/{id}/coach**
`GEMINI_API_KEY` not set, or model unavailable. Try `GEMINI_MODEL=gemini-1.5-flash`.

**CORS error in browser**
`CORS_ORIGINS` doesn't match the actual frontend domain (https vs http,
trailing slash, www subdomain). Echo back the exact `Origin` header your
browser sends and use that string.

**Postgres pool exhausted**
Railway's free Postgres caps at ~10 connections. The pgx pool defaults
to 4 — fine. If you scale the API to multiple replicas, lower the per-replica
pool or upgrade Postgres.
