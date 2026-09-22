# 82-0

Poké Draft is a Pokémon drafting game. This repository is a monorepo with a
Next.js frontend and a Go backend.

## Local development

Start the backend:

```powershell
cd backend
$env:DATABASE_URL = "postgres://pokemon_game:YOUR_PASSWORD@localhost:5432/pokemon_draft?sslmode=disable"
# Set this only after Redis is running locally:
# $env:REDIS_URL = "redis://localhost:6379/0"
go run ./cmd/api
```

Start the frontend in a separate terminal:

```powershell
cd frontend
Copy-Item .env.local.example .env.local
npm run dev
```

Open http://localhost:3000. The page calls the Go backend for its health status
and Pokémon data. The backend requests PokéAPI and calculates the displayed
base-stat score; the frontend does not calculate scores.

The backend requires a local PostgreSQL database. When a player completes a
six-Pokémon draft, the backend saves the server-calculated total as that
player's high score. Replace `YOUR_PASSWORD` with the password you created for
the `pokemon_game` PostgreSQL role. Do not put this connection string in the
frontend because it contains the database password.

Redis caching is optional but recommended. When `REDIS_URL` points to a running
Redis server, successful PokéAPI responses are cached for 24 hours. If Redis is
unavailable, the backend logs the problem and fetches from PokéAPI instead.

## Current API

`GET /api/pokemon/{name}` returns a Pokémon's six base stats and its calculated
score. For example: http://localhost:8080/api/pokemon/pikachu
