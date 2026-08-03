# Subscription Billing Platform

A modular billing platform built as a Go microservices sample project.

## Architecture

This repository includes three main services:

- **Subscription Service** (`services/subscription`)
  - GraphQL API for customers, plans, subscriptions, and invoices
  - JWT auth, Redis caching, Kafka event publishing
- **Payment Service** (`services/payment`)
  - gRPC payment processing service
  - PostgreSQL persistence and retry worker
- **Notification Service** (`services/notification`)
  - Kafka consumer for email notifications and event-driven messaging

Supporting infrastructure is defined in `docker-compose.yml` for local development.

## Security and Secrets

- Secrets are loaded from environment variables at runtime.
- Do not commit `.env` or `*.env` files to Git; they are ignored by `.gitignore`.
- Use `.env.example` as a template for local configuration.
- Rotate any leaked JWT secret or database password immediately.

## Local setup

1. Copy the root example file:

```bash
cp .env.example .env
```

2. Update `.env` with your own secure values:

- `POSTGRES_PASSWORD`
- `SECRET_KEY`
- `SECRET_REFRESH_KEY`

3. Start the whole stack locally:

```bash
docker compose up --build
```

4. Access the services:

- GraphQL playground: `http://localhost:8080`
- Payment service: `http://localhost:9090`
- Notification service: `http://localhost:8082`

## Run tests

Use Go test for individual service modules:

```bash
go test ./services/...
```

## Notes

- The project is configured for local development only.
- For production, use a secure secret store and enable TLS for database and gRPC connections.
- Keep actual credentials out of the repository by using env vars or Docker/Kubernetes secrets.
