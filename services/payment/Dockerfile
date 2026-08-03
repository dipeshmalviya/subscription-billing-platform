FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY gen /gen

COPY services/payment/go.mod services/payment/go.sum ./
RUN go mod download

COPY services/payment/. .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server


FROM alpine:3.19

# Install certificates and curl
RUN apk add --no-cache ca-certificates curl

# Install golang-migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/ && \
    chmod +x /usr/local/bin/migrate

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./db/migrations

EXPOSE 9090

CMD ["./server"]