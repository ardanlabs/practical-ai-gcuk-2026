# ==============================================================================
# Development

run:
	go run api/services/api/main.go

build:
	go build ./...

tidy:
	go mod tidy
	go mod vendor

# ==============================================================================
# Database

migrate:
	go run api/tooling/admin/main.go migrate

migrate-down:
	go run api/tooling/admin/main.go migrate-down

migrate-status:
	go run api/tooling/admin/main.go migrate-status

# ==============================================================================
# Testing

test:
	CGO_ENABLED=0 go test ./... -count=1
	go vet ./...

# test-curl drives the real HTTP API with curl. The service must be running,
# see the `run` target. Override the target with BASE_URL if needed.
test-curl:
	./zarf/integration/api_test.sh

.PHONY: run build tidy migrate migrate-down migrate-status test test-curl
