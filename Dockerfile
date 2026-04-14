FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /minima-mcp ./cmd/minima-mcp

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl && \
    addgroup -S minima && adduser -S minima -G minima

WORKDIR /app

COPY --from=builder /minima-mcp /app/minima-mcp
COPY --from=builder /build/.env.example /app/.env.example

RUN mkdir -p /app/audit-logs && chown -R minima:minima /app

USER minima

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:3001/health || exit 1

ENTRYPOINT ["/app/minima-mcp"]