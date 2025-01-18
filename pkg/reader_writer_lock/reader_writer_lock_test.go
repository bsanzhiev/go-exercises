package readerwriterlock

import (
	"sync"
	"testing"
	"time"
)

func TestTimeoutRWMutex_BasicLocking(t *testing.T) {
	mutex := NewTimeoutRWMutex()

	// Test basic write lock
	if !mutex.LockWithTimeout(time.Second) {
		t.Error("Failed to acquire write lock")
	}
	mutex.Unlock()

	// Test basic read lock
	if !mutex.RLockWithTimeout(time.Second) {
		t.Error("Failed to acquire read lock")
	}
	mutex.RUnlock()
}

func TestTimeoutRWMutex_Timeout(t *testing.T) {
	mutex := NewTimeoutRWMutex()

	// Acquire write lock first
	if !mutex.LockWithTimeout(time.Second) {
		t.Error("Failed to acquire initial write lock")
	}

	// Try to acquire another write lock (should timeout)
	if mutex.LockWithTimeout(100 * time.Millisecond) {
		t.Error("Second write lock should have timed out")
	}

	// Try to acquire read lock (should timeout)
	if mutex.RLockWithTimeout(100 * time.Millisecond) {
		t.Error("Read lock should have timed out")
	}

	mutex.Unlock()
}

func TestTimeoutRWMutex_MultipleReaders(t *testing.T) {
	mutex := NewTimeoutRWMutex()
	const readers = 5
	var wg sync.WaitGroup

	// Start multiple readers
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !mutex.RLockWithTimeout(time.Second) {
				t.Error("Failed to acquire read lock")
				return
			}
			time.Sleep(100 * time.Millisecond) // Simulate work
			mutex.RUnlock()
		}()
	}

	// Try to acquire write lock while readers are active (should timeout)
	if mutex.LockWithTimeout(50 * time.Millisecond) {
		t.Error("Write lock should have timed out while readers are active")
	}

	wg.Wait()
}

func TestTimeoutRWMutex_UnlockWithoutLock(t *testing.T) {
	mutex := NewTimeoutRWMutex()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on unlocking without lock")
		}
	}()

	mutex.Unlock()
}

func TestTimeoutRWMutex_RUnlockWithoutRLock(t *testing.T) {
	mutex := NewTimeoutRWMutex()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on unlocking without read lock")
		}
	}()

	mutex.RUnlock()
}
