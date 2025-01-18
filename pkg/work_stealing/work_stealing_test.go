package workstealing

import (
	"sync"
	"testing"
)

func TestWorkStealingPool(t *testing.T) {
	pool := NewWorkStealingPool(2)
	wg := sync.WaitGroup{}
	wg.Add(1)

	pool.Submit(func() {
		wg.Done()
	})

	wg.Wait()
}
