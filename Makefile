.PHONY: dev-db dev-backend dev-frontend dev-admin dev-stack dev-stack-down seed generate-types help

help:
	@echo "Available targets:"
	@echo "  db-start        - Start PostgreSQL via docker-compose"
	@echo "  dev-backend   - Run Go backend server"
	@echo "  dev-frontend  - Run vite server"
	@echo "  dev-admin     - Run admin frontend vite server (port 5174)"
	@echo "  dev-stack     - Run full stack in docker (backend, frontend, dev-auth) — open http://localhost:8090 to log in"
	@echo "  dev-stack-down - Tear down the dev stack"
	@echo "  check-frontend - Run vp check"
	@echo "  seed          - Run database migrations and seed data"
	@echo "  generate-types - Generate OpenAPI spec + shared TypeScript types"

db-start:
	docker-compose up -d

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && vp dev

dev-admin:
	cd admin && npm run dev

check-frontend:
	cd frontend && vp check

dev-stack:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

dev-stack-down:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml down

seed:
	cd backend && go run ./cmd/seed --migrate --all

generate-types:
	cd backend && go run ./cmd/openapi > ../shared/openapi.json
	cd frontend && vp exec openapi-typescript ../shared/openapi.json -o ../shared/api.generated.ts
