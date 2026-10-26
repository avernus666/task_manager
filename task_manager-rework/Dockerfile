FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/task-manager ./cmd/api

FROM alpine:3.21
RUN adduser -D -H -u 10001 appuser && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/task-manager /app/task-manager
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/task-manager"]
