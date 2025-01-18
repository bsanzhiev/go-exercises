package distributedcounter

import (
	"runtime"
	"sync"
	"testing"
)

func TestStripedCounter_Basic(t *testing.T) {
	counter := NewStripedCounter(runtime.GOMAXPROCS(0))

	counter.Increment()
	if got := counter.GetValue(); got != 1 {
		t.Errorf("After single increment, got %d, want 1", got)
	}
}

func TestStripedCounter_Concurrent(t *testing.T) {
	counter := NewStripedCounter(runtime.GOMAXPROCS(0))
	const numGoroutines = 100
	const incrementsPerGoroutine = 1000
	expected := uint64(numGoroutines * incrementsPerGoroutine)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				counter.Increment()
			}
		}()
	}

	wg.Wait()
	if got := counter.GetValue(); got != expected {
		t.Errorf("After concurrent increments, got %d, want %d", got, expected)
	}
}

func TestStripedCounter_DifferentStripes(t *testing.T) {
	tests := []struct {
		name       string
		numStripes int
		operations int
		goroutines int
		wantTotal  uint64
	}{
		{"Single Stripe", 1, 1000, 4, 4000},
		{"Two Stripes", 2, 1000, 4, 4000},
		{"Four Stripes", 4, 1000, 4, 4000},
		{"Eight Stripes", 8, 1000, 4, 4000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := NewStripedCounter(tt.numStripes)
			var wg sync.WaitGroup
			wg.Add(tt.goroutines)

			for i := 0; i < tt.goroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < tt.operations; j++ {
						counter.Increment()
					}
				}()
			}

			wg.Wait()
			if got := counter.GetValue(); got != tt.wantTotal {
				t.Errorf("got %d, want %d", got, tt.wantTotal)
			}
		})
	}
}

func BenchmarkStripedCounter(b *testing.B) {
	counter := NewStripedCounter(runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Increment()
		}
	})
}

func BenchmarkStripedCounter_GetValue(b *testing.B) {
	counter := NewStripedCounter(runtime.GOMAXPROCS(0))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.GetValue()
		}
	})
}
