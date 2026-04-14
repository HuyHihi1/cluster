package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// AsyncHandlerA publishes a message to NATS subject "async.a"
func AsyncHandlerA(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := otel.Tracer("handlers")

		ctx, span := tracer.Start(ctx, "AsyncHandlerA")
		defer span.End()

		msg := nats.NewMsg("async.a")
		msg.Data = []byte("Task A payload")

		// Inject the current span context into the NATS message headers
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))

		if err := nc.PublishMsg(msg); err != nil {
			slog.Error("failed to publish to NATS", "error", err)
			http.Error(w, "Failed to publish message", http.StatusInternalServerError)
			return
		}

		slog.Info("Published message to async.a", "trace_id", span.SpanContext().TraceID())
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "queued", "subject": "async.a"}`))
	}
}

// AsyncHandlerB publishes a message to NATS subject "async.b"
func AsyncHandlerB(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tracer := otel.Tracer("handlers")

		ctx, span := tracer.Start(ctx, "AsyncHandlerB")
		defer span.End()

		msg := nats.NewMsg("async.b")
		msg.Data = []byte("Task B payload")

		// Inject the current span context into the NATS message headers
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(msg.Header))

		if err := nc.PublishMsg(msg); err != nil {
			slog.Error("failed to publish to NATS", "error", err)
			http.Error(w, "Failed to publish message", http.StatusInternalServerError)
			return
		}

		slog.Info("Published message to async.b", "trace_id", span.SpanContext().TraceID())
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "queued", "subject": "async.b"}`))
	}
}

// StartWorker starts a NATS consumer that processes messages from "async.*"
func StartWorker(nc *nats.Conn) {
	slog.Info("Starting NATS worker...")

	_, err := nc.Subscribe("async.*", func(msg *nats.Msg) {
		// Extract the context from the message headers
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.HeaderCarrier(msg.Header))

		tracer := otel.Tracer("worker")
		ctx, span := tracer.Start(ctx, fmt.Sprintf("Process %s", msg.Subject), trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()

		slog.Info("Processing async task",
			"subject", msg.Subject,
			"data", string(msg.Data),
			"trace_id", span.SpanContext().TraceID(),
			"parent_span_id", trace.SpanContextFromContext(ctx).SpanID(),
		)

		// Simulate some background work
		time.Sleep(500 * time.Millisecond)

		slog.Info("Completed async task", "subject", msg.Subject)
	})

	if err != nil {
		slog.Error("failed to subscribe to NATS", "error", err)
	}
}
