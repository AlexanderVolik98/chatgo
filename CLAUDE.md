# ChatGo

Chat app: Go backend (`backend/`, net/http + PostgreSQL + WebSocket), Vue 3 frontend (`frontend/`, Vite + Pinia).
See README.md for architecture, API and WebSocket protocol.

## Commands

- `make up` — full stack in Docker (frontend :8081, backend :8080)
- `make backend-dev` — Postgres in Docker + backend locally
- `make frontend-dev` — Vite dev server on :5173
- `make seed` — load dev data (all users: `password123`)
- `cd backend && go test -race ./...` — backend tests

## Conventions

- Backend layers: `httpapi` (handlers) → `repository` (SQL) → `db`. Handlers depend on repository interfaces so they can be tested with in-memory mocks.
- Migrations live in `backend/internal/db/migrations`, embedded and applied on startup. Add new files, never edit applied ones.
- Standard library first; no web frameworks or ORMs.
- Comments are short, in English, and explain *why*, not *what*.

## Working with the AI assistant

The backend is written jointly with mutual review: the author reviews code written by the assistant, and the assistant reviews the author's code. When reviewing, point out problems and ask short "why" questions instead of rewriting.
