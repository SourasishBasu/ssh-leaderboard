# Stage 1: Install deps
FROM golang:1.23.6-bookworm AS deps

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Stage 2: Build application
FROM golang:1.23.6-bookworm AS builder

WORKDIR /app

COPY --from=deps /go/pkg /go/pkg
COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN mkdir -p bin/ .ssh/
RUN go build -ldflags="-w -s" -o bin/leaderboard cmd/ssh-leaderboard/ssh-leaderboard.go

# Stage 3: Run binary
FROM debian:bookworm-slim

WORKDIR /app

RUN groupadd -r appuser && useradd -r -g appuser appuser

COPY --from=builder /app/bin/leaderboard ./leaderboard
COPY --from=builder /app/.ssh ./.ssh
COPY --from=builder /app/cmd/ssh-leaderboard/.env ./.env

RUN chown appuser:appuser /app

EXPOSE 23234

USER appuser

CMD ["/app/leaderboard"]