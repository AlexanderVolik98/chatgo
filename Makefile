.PHONY: up down logs build backend-dev frontend-dev seed

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

build:
	docker compose build

db-up: 
	docker compose up -d postgres

backend-dev: db-up
	cd backend && DATABASE_URL="postgres://chatgo:chatgo@localhost:5432/chatgo?sslmode=disable" JWT_SECRET="dev-secret" go run ./cmd/server

frontend-dev:
	cd frontend && npm install && npm run dev

seed: db-up
	docker compose exec -T postgres psql -U chatgo -d chatgo < backend/internal/db/seed.sql
