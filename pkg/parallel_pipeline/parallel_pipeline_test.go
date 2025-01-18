package parallelpipeline

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Helper function to print channel results
func printResults(t *testing.T, ch <-chan interface{}) {
	for result := range ch {
		t.Logf("Result: %v", result)
	}
}

// Test processors
func multiply(x interface{}) (interface{}, error) {
	num := x.(int)
	return num * 2, nil
}

func addOne(x interface{}) (interface{}, error) {
	num := x.(int)
	return num + 1, nil
}

func TestPipelineWithPrinting(t *testing.T) {
	// Create stages
	stage1 := ParallelStage(2, 10, multiply)
	stage2 := ParallelStage(2, 10, addOne)

	// Create pipeline
	pipeline := NewPipeline(stage1, stage2)
	defer pipeline.Stop()

	// Create input channel
	input := make(chan interface{}, 10)

	// Start pipeline
	output := pipeline.Run(input)

	// Send test data
	go func() {
		for i := 1; i <= 5; i++ {
			input <- i
		}
		close(input)
	}()

	// Print results
	printResults(t, output)
}

func TestPipelineWithTimeout(t *testing.T) {
	// Create a slow processor
	slowProcessor := func(x interface{}) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		num := x.(int)
		return num * 2, nil
	}

	// Create pipeline with slow stage
	pipeline := NewPipeline(
		ParallelStage(2, 5, slowProcessor),
		ParallelStage(2, 5, addOne),
	)
	defer pipeline.Stop()

	// Create input channel
	input := make(chan interface{}, 5)

	// Start pipeline with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	output := pipeline.Run(input)

	// Send test data
	go func() {
		for i := 1; i <= 10; i++ {
			select {
			case <-ctx.Done():
				close(input)
				return
			case input <- i:
			}
		}
		close(input)
	}()

	// Print results with timeout
	for {
		select {
		case <-ctx.Done():
			t.Log("Pipeline timeout")
			return
		case result, ok := <-output:
			if !ok {
				return
			}
			t.Logf("Result: %v", result)
		}
	}
}

// Example of usage that can be run as go test -v
func ExamplePipeline() {
	// Create stages
	stage1 := ParallelStage(2, 5, multiply)
	stage2 := ParallelStage(2, 5, addOne)

	// Create pipeline
	pipeline := NewPipeline(stage1, stage2)
	defer pipeline.Stop()

	// Create input channel
	input := make(chan interface{}, 5)

	// Start pipeline
	output := pipeline.Run(input)

	// Send test data
	go func() {
		for i := 1; i <= 3; i++ {
			input <- i
		}
		close(input)
	}()

	// Print results
	for result := range output {
		fmt.Printf("Result: %v\n", result)
	}

	// Output:
	// Result: 3
	// Result: 5
	// Result: 7
}
