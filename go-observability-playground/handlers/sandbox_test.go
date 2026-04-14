package handlers

import (
	"context"
	"testing"
)

// BenchmarkCompute function measures the performance of the calculateFactorial helper.
// Run with: go test -bench=BenchmarkCompute -v
func BenchmarkCompute(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer() // Reset timer before the loop
	for i := 0; i < b.N; i++ {
		// Try a smaller factorial so the benchmark runs reasonably quickly
		_ = calculateFactorial(ctx, 1000)
	}
}

// BenchmarkMemory function measures memory allocation task
// Run with: go test -bench=BenchmarkMemory -v -benchmem
func BenchmarkMemory(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate 5mb in loop
		_ = allocateMemory(ctx, 5)
	}
}
