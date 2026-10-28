# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/task-manager ./cmd/api

FROM migrate/migrate:v4.18.2 AS migrate-cli

FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=build /out/task-manager /app/task-manager
COPY --from=migrate-cli /usr/local/bin/migrate /usr/local/bin/migrate
COPY db/migrations /app/migrations
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod 0555 /app/docker-entrypoint.sh /usr/local/bin/migrate \
    && chown -R appuser:appuser /app
USER appuser
ENV HTTP_ADDRESS=:8080
EXPOSE 8080
ENTRYPOINT ["/app/docker-entrypoint.sh"]
