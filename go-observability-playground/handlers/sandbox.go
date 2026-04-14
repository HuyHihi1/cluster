package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"go-observability-playground/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("sandbox-handler")

// ComputeHandler simulates CPU intensive work (calculating factorials).
// This is heavily unoptimized to show up prominently in pprof CPU profiles.
func ComputeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "ComputeHandler")
	defer span.End()

	slog.InfoContext(ctx, "Starting compute-intensive task")
	observability.ComputeTasksProcessed.Inc()

	n := 50000 // Compute factorial of 50,000
	result := calculateFactorial(ctx, n)

	span.SetAttributes(attribute.Int("compute.n", n))
	span.SetAttributes(attribute.Int("compute.result_length", len(result.String())))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"task":    "factorial",
		"n":       n,
		"success": true,
	})
}

// MemoryHandler simulates memory intensive work (large un-gc'd slice allocations).
func MemoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "MemoryHandler")
	defer span.End()

	slog.InfoContext(ctx, "Starting memory-intensive task")
	observability.MemoryAllocations.Inc()

	// Allocate 100MB of slices
	allocateMemory(ctx, 100)

	span.SetAttributes(attribute.Int("memory.allocated_mb", 100))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"task":    "memory_allocation",
		"size_mb": 100,
		"success": true,
	})
}

// MockDBHandler simulates a slow database query. Useful for tracing visualization!
func MockDBHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "MockDBHandler")
	defer span.End()

	slog.InfoContext(ctx, "Starting simulated user fetch from db")

	// Span: Validating Request
	_, valSpan := tracer.Start(ctx, "ValidateRequest")
	time.Sleep(5 * time.Millisecond)
	valSpan.End()

	// Span: Query Database
	dbCtx, dbSpan := tracer.Start(ctx, "QueryDatabase")
	dbSpan.SetAttributes(attribute.String("db.system", "postgresql"))
	dbSpan.SetAttributes(attribute.String("db.statement", "SELECT * FROM users WHERE active = true"))

	// Simulate db latency
	time.Sleep(150 * time.Millisecond)

	slog.DebugContext(dbCtx, "Executed SQL query")
	dbSpan.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"users_fetched": 42,
		"duration_ms":   155,
	})
}

// Helpers

func calculateFactorial(ctx context.Context, n int) *big.Int {
	_, span := tracer.Start(ctx, "calculateFactorial")
	defer span.End()

	fact := big.NewInt(1)
	for i := 1; i <= n; i++ {
		fact.Mul(fact, big.NewInt(int64(i)))
	}
	return fact
}

func allocateMemory(ctx context.Context, sizeMB int) [][]byte {
	_, span := tracer.Start(ctx, "allocateMemory")
	defer span.End()

	var data [][]byte
	for i := 0; i < sizeMB; i++ {
		mb := make([]byte, 1024*1024)
		// fill to prevent compiler optimizations
		for j := range mb {
			mb[j] = byte(i % 256)
		}
		data = append(data, mb)
	}
	// sleep briefly to hold memory and let profile capture it
	time.Sleep(100 * time.Millisecond)
	return data
}
