package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// NatsCarrier satisfies the propagation.TextMapCarrier interface
type NatsCarrier struct {
	Header nats.Header
}

func (c NatsCarrier) Get(key string) string {
	return c.Header.Get(key)
}

func (c NatsCarrier) Set(key, value string) {
	c.Header.Set(key, value)
}

func (c NatsCarrier) Keys() []string {
	keys := make([]string, 0, len(c.Header))
	for k := range c.Header {
		keys = append(keys, k)
	}
	return keys
}

// TriggerAsyncHandler publishes a message to NATS subject "tasks.process"
func TriggerAsyncHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := otel.Tracer("server-a")

		ctx, span := tracer.Start(ctx, "TriggerAsyncHandler")
		defer span.End()

		payload := fmt.Sprintf("Task triggered at %s", time.Now().Format(time.RFC3339))
		msg := nats.NewMsg("tasks.process")
		msg.Data = []byte(payload)
		msg.Header = make(nats.Header)

		// Inject the current span context using our NatsCarrier
		otel.GetTextMapPropagator().Inject(ctx, NatsCarrier{Header: msg.Header})

		if err := nc.PublishMsg(msg); err != nil {
			slog.Error("failed to publish to NATS", "error", err)
			http.Error(w, "Failed to publish message", http.StatusInternalServerError)
			return
		}

		slog.Info("Published message to tasks.process", "trace_id", span.SpanContext().TraceID())
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status": "queued", "payload": "%s", "trace_id": "%s"}`, payload, span.SpanContext().TraceID())))
	}
}

// StartServerBWorker starts a NATS consumer that processes messages and saves to Redis
func StartServerBWorker(nc *nats.Conn, rdb *redis.Client) {
	slog.Info("Starting Server B NATS worker...")

	// Enable tracing for go-redis
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		slog.Error("failed to instrument redis with tracing", "error", err)
	}

	_, err := nc.Subscribe("tasks.process", func(msg *nats.Msg) {
		// Extract the context using our NatsCarrier
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), NatsCarrier{Header: msg.Header})

		tracer := otel.Tracer("server-b")
		// Start a new span as a child of the extracted context
		ctx, span := tracer.Start(ctx, "ProcessTask", trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()

		slog.Info("Processing task in Server B",
			"data", string(msg.Data),
			"trace_id", span.SpanContext().TraceID(),
		)

		// Simulate some processing time
		time.Sleep(200 * time.Millisecond)

		// Save to Redis (using the new ctx with the correct TraceID)
		key := fmt.Sprintf("task:%s", span.SpanContext().TraceID())
		err := rdb.Set(ctx, key, string(msg.Data), 1*time.Hour).Err()
		if err != nil {
			slog.Error("failed to save to Redis", "error", err)
			return
		}

		slog.Info("Successfully saved task to Redis", "key", key)
	})

	if err != nil {
		slog.Error("failed to subscribe to NATS", "error", err)
	}
}
