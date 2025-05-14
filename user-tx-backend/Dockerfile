# ─── Build Stage ───────────────────────────────────────────────────────────────
FROM golang:1.20-alpine AS builder
WORKDIR /app

# Grab go.mod and go.sum, then download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy everything and build statically
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o user-tx-backend ./main.go

# ─── Final Stage ───────────────────────────────────────────────────────────────
FROM alpine:latest
RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /app/user-tx-backend .

# Expose the port configured in code (default 8080)
EXPOSE 8080

COPY wait-for-neo4j.sh .

# The binary will read NEO4J_URI, NEO4J_USER, NEO4J_PASS, SEED_DATA, PORT from env
ENTRYPOINT ["./wait-for-neo4j.sh"]
