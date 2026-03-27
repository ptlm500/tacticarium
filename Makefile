.PHONY: dev-db dev-backend dev-frontend seed help

help:
	@echo "Available targets:"
	@echo "  dev-db        - Start PostgreSQL via docker-compose"
	@echo "  dev-backend   - Run Go backend server"
	@echo "  dev-frontend  - Run React dev server"
	@echo "  seed          - Run database migrations and seed data"

dev-db:
	docker-compose up -d

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

seed:
	cd backend && go run ./cmd/seed --migrate --all
