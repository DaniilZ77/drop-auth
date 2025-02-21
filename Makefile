SERVICE_NAME=drop-auth

TEST_FLAGS=-count=1
BUILD_FLAGS=

.PHONY: run, build, lint, test, coverage, migrate-new, migrate-up, migrate-down

run: ### run app
	CONFIG_PATH=./config/local.yaml go run cmd/auth/main.go

build: ### build app
	go build ${BUILD_FLAGS} -o ${SERVICE_NAME} cmd/auth/main.go

lint: ### run linter
	@golangci-lint --timeout=2m run

test: ### run test
	go test ${TEST_FLAGS} ./...

coverage: ### generate coverage report
	go test ${TEST_FLAGS} -coverprofile=coverage.out ./...
	go tool cover -html="coverage.out"

MIGRATION_NAME=initial

migrate-new: ### create a new migration
	migrate create -ext sql -dir ./internal/db/migrations -seq ${MIGRATION_NAME}

migrate-up: ### apply all migrations
	migrate -path ./internal/db/migrations -database '$(PG_URL)?sslmode=disable' up

migrate-down: ### migration down
	migrate -path ./internal/db/migrations -database '$(PG_URL)?sslmode=disable' down