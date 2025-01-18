package ringbuffer

import (
	"sync"
	"testing"
)

func TestRingBufferBasic(t *testing.T) {
	rb := NewRingBuffer(2)
	if err := rb.Put("A"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	item, err := rb.Get()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item != "A" {
		t.Errorf("expected 'A', got %v", item)
	}
}

func TestRingBufferConcurrent(t *testing.T) {
	rb := NewRingBuffer(5)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			_ = rb.Put(val)
		}(i)
	}
	wg.Wait()
	for i := 0; i < 5; i++ {
		item, err := rb.Get()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if item != i {
			t.Errorf("expected %d, got %v", i, item)
		}
	}
}

func TestRingBufferBlocking(t *testing.T) {
	rb := NewRingBuffer(1)
	done := make(chan struct{})
	go func() {
		_ = rb.Put("A")
		close(done)
	}()
	<-done

	go func() {
		_ = rb.Put("B")
		close(done)
	}()
	select {
	case <-done:
		t.Errorf("expected blocking on full buffer")
	default:
	}

	go func() {
		_, _ = rb.Get()
		close(done)
	}()
	select {
	case <-done:
	default:
		t.Errorf("expected blocking on empty buffer")
	}
}

func TestRingBufferClose(t *testing.T) {
	rb := NewRingBuffer(1)
	_ = rb.Put("A")
	rb.mu.Lock()
	rb.closed = true
	rb.mu.Unlock()

	if err := rb.Put("B"); err == nil {
		t.Errorf("expected error on closed buffer")
	}

	_, err := rb.Get()
	if err == nil {
		t.Errorf("expected error on closed buffer")
	}
}
