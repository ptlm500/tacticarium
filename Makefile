.PHONY: dev-db dev-backend dev-frontend seed help

help:
	@echo "Available targets:"
	@echo "  db-start        - Start PostgreSQL via docker-compose"
	@echo "  dev-backend   - Run Go backend server"
	@echo "  dev-frontend  - Run vite server"
	@echo "  check-frontend - Run vp check"
	@echo "  seed          - Run database migrations and seed data"

db-start:
	docker-compose up -d

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && vp dev

check-frontend:
	cd frontend && vp check

seed:
	cd backend && go run ./cmd/seed --migrate --all
