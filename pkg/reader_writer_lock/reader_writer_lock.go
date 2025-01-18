package readerwriterlock

import (
	"sync"
	"time"
)

/*
Reader-Writer Lock with Timeout:
Реализует справедливую схему доступа
Поддерживает таймауты для операций
Предотвращает starving писателей
Использует условные переменные для синхронизации
*/

/*
Условие:
Реализовать RWMutex с поддержкой timeout для операций блокировки.
Требования:
- Поддержка timeout для операций чтения и записи
- Приоритет писателей над читателями
- Отсутствие starving писателей
- Корректная работа с контекстом
*/

type TimeoutRWMutex struct {
	mu          sync.Mutex
	readers     int
	writing     bool
	writeWait   *sync.Cond
	readWait    *sync.Cond
	writerCount int
}

func NewTimeoutRWMutex() *TimeoutRWMutex {
	trw := &TimeoutRWMutex{}
	trw.writeWait = sync.NewCond(&trw.mu)
	trw.readWait = sync.NewCond(&trw.mu)
	return trw
}

func (trw *TimeoutRWMutex) LockWithTimeout(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	trw.mu.Lock()
	// Increment waiting writers count
	trw.writerCount++

	// Wait until no readers and no active writer
	for trw.writing || trw.readers > 0 {
		done := make(chan struct{})
		go func() {
			trw.writeWait.Wait()
			close(done)
		}()

		trw.mu.Unlock()
		select {
		case <-done:
			trw.mu.Lock()
			continue
		case <-timer.C:
			trw.mu.Lock()
			trw.writerCount--
			trw.readWait.Broadcast()
			trw.mu.Unlock()
			return false
		}
	}

	trw.writing = true
	trw.writerCount--
	trw.mu.Unlock()
	return true
}

func (trw *TimeoutRWMutex) RLockWithTimeout(timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	trw.mu.Lock()
	// Wait while there's an active writer or waiting writers
	for trw.writing || trw.writerCount > 0 {
		done := make(chan struct{})
		go func() {
			trw.readWait.Wait()
			close(done)
		}()

		trw.mu.Unlock()
		select {
		case <-done:
			trw.mu.Lock()
			continue
		case <-timer.C:
			trw.mu.Lock()
			trw.mu.Unlock()
			return false
		}
	}

	trw.readers++
	trw.mu.Unlock()
	return true
}

func (trw *TimeoutRWMutex) Unlock() {
	trw.mu.Lock()
	trw.writing = false
	trw.writeWait.Broadcast()
	trw.readWait.Broadcast()
	trw.mu.Unlock()
}

func (trw *TimeoutRWMutex) RUnlock() {
	trw.mu.Lock()
	trw.readers--
	if trw.readers == 0 {
		trw.writeWait.Broadcast()
	}
	trw.mu.Unlock()
}
