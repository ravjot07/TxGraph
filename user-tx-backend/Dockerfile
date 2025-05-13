# ─── Build Stage ─────────────────────────────────────────────────────────────
FROM golang:1.20-alpine AS builder
WORKDIR /app

# Cache deps
COPY go.mod go.sum ./
RUN go mod tidy

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ─── Run Stage ───────────────────────────────────────────────────────────────
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
