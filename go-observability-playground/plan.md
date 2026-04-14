# Multi-Hop Async Tracing Pipeline Plan

The goal is to demonstrate a complex, multi-stage distributed trace using NATS as the event bus and Redis as the state store.

## User Review Required

> [!IMPORTANT]
>
> - I will add **Redis** as a new service in `docker-compose.yml`.
> - I will install `github.com/redis/go-redis/v9` for Go connectivity.
> - The application will grow to include multiple "micro-workers" (goroutines) representing different stages of an order lifecycle.

## Proposed Changes

### Infrastructure

#### [MODIFY] [docker-compose.yml](file:///Users/qhuy/Documents/study/go-observability-playground/docker-compose.yml)

- Add a `redis` service.
- Update `web-app` environment to include `REDIS_URL`.

### Go Application

#### [MODIFY] [go.mod](file:///Users/qhuy/Documents/study/go-observability-playground/go.mod)

- Add `github.com/redis/go-redis/v9` and `github.com/redis/go-redis/extra/redisotel/v9` for automatic Redis tracing.

#### [NEW] [pipeline.go](file:///Users/qhuy/Documents/study/go-observability-playground/handlers/pipeline.go)

- **Order Handler**: HTTP entry point that saves to Redis and publishes `order.requested`.
- **Payment Worker**: Consumes `order.requested`, simulates processing, and publishes `order.paid`.
- **Fulfillment Worker**: Consumes `order.paid`, simulates shipping, and publishes `order.completed`.
- **Notification Worker**: Consumes `order.completed` and finalizes the trace.

#### [MODIFY] [main.go](file:///Users/qhuy/Documents/study/go-observability-playground/main.go)

- Initialize Redis connection with OTel instrumentation.
- Register endpoints and start all pipeline workers.

## Open Questions

- Should I expose this via a new route like `/api/v1/order/create` or replace the existing async test? (Plan: Add alongside as a "Complex Flow").
- Do you want to see any specific metrics (Prometheus) tied to these stages? (Plan: Will add standard OTel metrics).

## Verification Plan

### Automated Verification

- Run `go mod tidy` and `go build`.
- Use the `browser` tool to verify the Tempo waterfall in Grafana.
- Verify that Redis interactions show up as sub-spans within the worker spans.

### Manual Verification

- Check logs for the unique Trace ID propagating through all 4 stages.
