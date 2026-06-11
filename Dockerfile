FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-server ./cmd/server/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/wallet-server .
COPY example.config.env ./config.env
COPY migrations ./migrations

EXPOSE 8080

CMD ["./wallet-server"]
