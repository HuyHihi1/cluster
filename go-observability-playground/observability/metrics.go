package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCount tracks all incoming HTTP requests
	RequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	// ComputeTasksProcessed tracks the number of times the CPU-intensive /compute endpoint is hit
	ComputeTasksProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "compute_tasks_total",
			Help: "Total number of compute tasks initiated",
		},
	)

	// MemoryAllocations tracks memory-intensive requests
	MemoryAllocations = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "memory_allocation_tasks_total",
			Help: "Total number of memory allocation tasks triggered",
		},
	)
)
