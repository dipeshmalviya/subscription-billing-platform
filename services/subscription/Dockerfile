FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY gen /gen
COPY services/subscription/go.mod services/subscription/go.sum ./
RUN go mod download

COPY services/subscription/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server


FROM alpine:3.19

# Install required packages
RUN apk add --no-cache ca-certificates curl

# Install golang-migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz \
    | tar -xz

RUN mv migrate /usr/local/bin/

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/db/migrations ./db/migrations

EXPOSE 8080

CMD ["./server"]