package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-observability-playground/handlers"
	"go-observability-playground/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// 1. Setup Structured Logger
	observability.InitLogger()

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "go-observability-playground"
	}

	// 2. Setup OpenTelemetry Distributed Tracing
	ctx := context.Background()
	tp, err := observability.InitTracer(ctx, serviceName)
	if err != nil {
		slog.Error("failed to init tracer", "error", err)
	} else {
		defer func() {
			if err := tp.Shutdown(ctx); err != nil {
				slog.Error("failed to shutdown TracerProvider", "error", err)
			}
		}()
	}

	// 3. Connect to NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		slog.Error("failed to connect to NATS", "url", natsURL, "error", err)
	} else {
		defer nc.Close()
		slog.Info("Connected to NATS", "url", natsURL)
	}

	// 4. Connect to Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer rdb.Close()
	slog.Info("Connected to Redis", "url", redisURL)

	// 5. Start Workers if applicable
	if serviceName == "server-b" && nc != nil {
		handlers.StartServerBWorker(nc, rdb)
	}

	// 6. Setup Chi Router
	r := chi.NewRouter()

	// Chi Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Application custom Prometheus metrics middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			
			// Extract matched route pattern rather than exact URI (to prevent high cardinality)
			ctx := chi.RouteContext(r.Context())
			pattern := "unknown"
			if ctx != nil && ctx.RoutePattern() != "" {
				pattern = ctx.RoutePattern()
			}
			observability.RequestCount.WithLabelValues(r.Method, pattern, http.StatusText(ww.Status())).Inc()
		})
	})

	// 7. Exposed Standard Endpoints (Metrics and PProf Profiling)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"service": "` + serviceName + `", "message": "Welcome to the Go Observability Playground!",
		 "endpoints": ["/api/v1/trigger-async", "/api/v1/compute", "/api/v1/memory", "/api/v1/mock-db", "/metrics", "/debug/pprof/"]}`))
	})
	r.Mount("/metrics", promhttp.Handler())

	r.Route("/debug/pprof", func(r chi.Router) {
		r.Get("/", pprof.Index)
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Handle("/allocs", pprof.Handler("allocs"))
		r.Handle("/block", pprof.Handler("block"))
		r.Handle("/goroutine", pprof.Handler("goroutine"))
		r.Handle("/heap", pprof.Handler("heap"))
		r.Handle("/mutex", pprof.Handler("mutex"))
		r.Handle("/threadcreate", pprof.Handler("threadcreate"))
	})

	// 8. Application API Endpoints (wrapped in OTel Instrumentation)
	r.Route("/api/v1", func(api chi.Router) {
		// Use OTel http middleware around all API endpoints
		api.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewHandler(next, "api/v1")
		})

		api.Get("/compute", handlers.ComputeHandler)
		api.Get("/memory", handlers.MemoryHandler)
		api.Get("/mock-db", handlers.MockDBHandler)

		if serviceName == "server-a" && nc != nil {
			api.Get("/trigger-async", handlers.TriggerAsyncHandler(nc))
		}
	})

	// Graceful Shutdown
	port := ":8080"
	if serviceName == "server-b" {
		port = ":8081" // Use different port for local testing if needed, though in docker it doesn't matter much
	}

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		slog.Info("Starting server", "service", serviceName, "addr", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for an interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down server...", "service", serviceName)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	}
	slog.Info("Server exited gracefully", "service", serviceName)
}
