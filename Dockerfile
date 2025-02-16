FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/drop-auth ./cmd/auth
RUN go build -o bin/migrator ./cmd/migrator

FROM alpine:3.21
RUN apk --no-cache add curl
WORKDIR /app
COPY --from=builder ./app/bin ./bin
COPY --from=builder ./app/internal/db/migrations ./migrations
COPY --from=builder ./app/tls ./tls