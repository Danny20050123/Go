# 82-0

Poké Draft is a Pokémon drafting game. This repository is a monorepo with a
Next.js frontend and a Go backend.

## Local development

Start the backend:

```powershell
cd backend
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

## Current API

`GET /api/pokemon/{name}` returns a Pokémon's six base stats and its calculated
score. For example: http://localhost:8080/api/pokemon/pikachu
