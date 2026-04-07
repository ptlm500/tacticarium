.PHONY: dev-db dev-backend dev-frontend dev-admin seed generate-types help

help:
	@echo "Available targets:"
	@echo "  db-start        - Start PostgreSQL via docker-compose"
	@echo "  dev-backend   - Run Go backend server"
	@echo "  dev-frontend  - Run vite server"
	@echo "  dev-admin     - Run admin frontend vite server (port 5174)"
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

seed:
	cd backend && go run ./cmd/seed --migrate --all

generate-types:
	cd backend && go run ./cmd/openapi > ../shared/openapi.json
	cd frontend && vp exec openapi-typescript ../shared/openapi.json -o ../shared/api.generated.ts
