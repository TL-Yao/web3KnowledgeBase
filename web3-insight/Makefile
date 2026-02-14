.PHONY: dev dev-backend dev-frontend db-up db-down test build clean

# Development
dev: db-up
	@make -j2 dev-backend dev-frontend

dev-backend:
	/usr/local/go/bin/go run -C backend cmd/server/main.go

dev-frontend:
	cd frontend && npm run dev

# Database
db-up:
	docker-compose up -d postgres redis
	@echo "Waiting for database..."
	@sleep 3

db-down:
	docker-compose down

# Worker
worker:
	/usr/local/go/bin/go run -C backend cmd/worker/main.go

# Build
build-backend:
	/usr/local/go/bin/go build -C backend -o bin/server cmd/server/main.go
	/usr/local/go/bin/go build -C backend -o bin/worker cmd/worker/main.go

build-frontend:
	cd frontend && npm run build

build: build-backend build-frontend

# Test
test-backend:
	/usr/local/go/bin/go test -C backend ./...

test-frontend:
	cd frontend && npm test

test: test-backend test-frontend

# Clean
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules
