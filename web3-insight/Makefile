.PHONY: dev dev-backend dev-frontend db-up db-down migrate seed test build clean

# Development
dev: db-up
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && go run cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

# Database
db-up:
	docker-compose up -d postgres redis
	@echo "Waiting for database..."
	@sleep 3

db-down:
	docker-compose down

migrate:
	cd backend && go run cmd/migrate/main.go

seed:
	cd backend && go run cmd/seed/main.go

# Worker
worker:
	cd backend && go run cmd/worker/main.go

# Build
build-backend:
	cd backend && go build -o bin/server cmd/server/main.go
	cd backend && go build -o bin/worker cmd/worker/main.go

build-frontend:
	cd frontend && npm run build

build: build-backend build-frontend

# Test
test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

test: test-backend test-frontend

# Clean
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules
