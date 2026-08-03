FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY services/notification/go.mod services/notification/go.sum ./
RUN go mod download

COPY services/notification/. .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8082
CMD ["./server"]