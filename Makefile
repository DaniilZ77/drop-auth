SERVICE_NAME=drop-auth

TEST_FLAGS=-count=1
BUILD_FLAGS=

.PHONY: run, build, lint, test, coverage, migrate-new, migrate-up, migrate-down

# TODO define your envs, switch log_level to `debug` during developing
PG_URL=postgres://postgres:postgres@localhost:5432/drop-auth

run: ### run app
	go run cmd/auth/main.go -db_url '$(PG_URL)' \
	-grpc_port localhost:50051 -tma_secret 5768337691:AAH5YkoiEuPk8-FZa32hStHTqXiLPtAEhx8 \
	-access_token_ttl 15 -refresh_token_ttl 14400 -read_timeout 5 \
	-http_port localhost:8080 -env local -cert ./tls/cert.pem \
	-key ./tls/key.pem -jwt_secret secret -redis_addr localhost:6379

build: ### build app
	go build ${BUILD_FLAGS} -o ${SERVICE_NAME} cmd/auth/main.go

lint: ### run linter
	@golangci-lint --timeout=2m run

test: ### run test
	go test ${TEST_FLAGS} ./...

coverage: ### generate coverage report
	go test ${TEST_FLAGS} -coverprofile=coverage.out ./...
	go tool cover -html="coverage.out"

MIGRATION_NAME=auth-tg

migrate-new: ### create a new migration
	migrate create -ext sql -dir ./internal/db/migrations -seq ${MIGRATION_NAME}

migrate-up: ### apply all migrations
	migrate -path ./internal/db/migrations -database '$(PG_URL)?sslmode=disable' up

migrate-down: ### migration down
	migrate -path ./internal/db/migrations -database '$(PG_URL)?sslmode=disable' down

mock:
	mockery