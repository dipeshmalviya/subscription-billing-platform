# Subscription Billing Platform

## Objective

Build a production-inspired Subscription Billing System using Go microservices.

## What this project delivers

- **GraphQL API** for customer, subscription, invoice, and billing management.
- **Subscription Service** and **Payment Service** communicating via gRPC.
- **PostgreSQL** persistence with normalized schema, transactions, and worker support.
- **Redis** caching for plans and customer summaries.
- **Kafka** event publishing with a Notification Service consumer.
- **JWT authentication** and role-based authorization for Admin and Customer flows.
- **Background workers** for subscription renewals and payment retries.
- **Structured logging**, health checks, and Prometheus metrics.
- **Docker Compose** for local development and Kubernetes manifests for deployment.

## Architecture overview

The system is organized into three main services:

- **Subscription Service** (`services/subscription`)
  - GraphQL server for customer signup/login, plans, subscription lifecycle, and invoices.
  - JWT-based auth middleware and role enforcement.
  - Redis caching for fast plan and summary reads.
  - Publishes subscription and invoice events to Kafka.
- **Payment Service** (`services/payment`)
  - gRPC payment processing API.
  - Handles payment creation, retries, and persistence in PostgreSQL.
  - Exposes metrics and health endpoints.
- **Notification Service** (`services/notification`)
  - Kafka consumer that sends notification events.
  - Designed for email or webhook notification delivery.

### Infrastructure

- PostgreSQL for relational storage.
- Redis for caching.
- Kafka for event-driven communication.
- Docker Compose for local orchestration.
- Kubernetes manifests in `deploy/k8s/` for production-style deployment.
- Architecture diagram in `architecture/architecture-diagram.png`.

## Repo structure

- `services/subscription/`
- `services/payment/`
- `services/notification/`
- `docker-compose.yml`
- `deploy/k8s/`
- `architecture/architecture-diagram.png`

## Setup and run locally

### 1. Copy environment examples

```bash
cp .env.example .env
cp services/payment/.env.example services/payment/.env
cp services/subscription/.env.example services/subscription/.env
```

### 2. Update secrets

Open the copied `.env` files and replace placeholders with secure values.

Required fields:

- `POSTGRES_PASSWORD`
- `SECRET_KEY`
- `SECRET_REFRESH_KEY`
- `REDIS_ADDR`
- `KAFKA_BROKERS`
- `PAYMENT_SERVICE_ADDR`

### 3. Run the application stack

```bash
docker compose up --build
```

### 4. Visit the services

- GraphQL playground: `http://localhost:8080`
- Payment service: `http://localhost:9090`
- Notification service: `http://localhost:8082`

## Sample GraphQL queries

### Signup

```graphql
mutation Signup($input: SignupInput!) {
  signup(input: $input) {
    accessToken
    refreshToken
    customer {
      id
      email
      fullName
      role
    }
  }
}
```

### Login

```graphql
mutation Login($input: LoginInput!) {
  login(input: $input) {
    accessToken
    refreshToken
  }
}
```

### Fetch plans

```graphql
query Plans {
  plans {
    id
    name
    priceCents
    interval
    description
  }
}
```

### Create subscription

```graphql
mutation CreateSubscription($input: CreateSubscriptionInput!) {
  createSubscription(input: $input) {
    id
    planId
    customerId
    status
    startsAt
    endsAt
  }
}
```

## Running tests

Run Go tests for all services:

```bash
go test ./services/...
```

## Kubernetes manifests

Service deployment manifests are available in `deploy/k8s/`.
Use those manifests when moving the stack to Kubernetes.

## Security and best practices

- Keep actual credentials out of Git.
- Use `.env.example` templates only.
- Rotate JWT secrets and DB passwords if leaked.
- Use a secret manager or Kubernetes secrets for production.
- Enable TLS for database and gRPC connections in production.

## Notes

This repository is intended as a developer-friendly sample billing platform. It is structured for local development, with an easy path to Docker Compose and Kubernetes deployment.
