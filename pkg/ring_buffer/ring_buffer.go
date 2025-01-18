package ringbuffer

import (
	"fmt"
	"sync"
)

/*
Ring Buffer Queue:
Использует условные переменные для синхронизации
Эффективно использует память за счет кольцевой структуры
Поддерживает множество производителей и потребителей
Имеет защиту от гонок данных
*/

/*
Условие:
Реализовать потокобезопасную очередь на кольцевом буфере фиксированного размера.
Требования:
- Поддержка множества производителей и потребителей
- Блокировка при полном буфере/пустой очереди
- Корректная работа с условными переменными
- Возможность безопасного завершения работы
*/

type RingBuffer struct {
	buffer   []interface{}
	size     int
	readIdx  int
	writeIdx int
	count    int
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	closed   bool
}

func NewRingBuffer(size int) *RingBuffer {
	rb := &RingBuffer{
		buffer: make([]interface{}, size),
		size:   size,
	}
	rb.notEmpty = sync.NewCond(&rb.mu)
	rb.notFull = sync.NewCond(&rb.mu)
	return rb
}

func (rb *RingBuffer) Put(item interface{}) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return fmt.Errorf("buffer is closed")
	}

	for rb.count == rb.size {
		if rb.closed {
			return fmt.Errorf("buffer is closed")
		}
		rb.notFull.Wait()
	}

	rb.buffer[rb.writeIdx] = item
	rb.writeIdx = (rb.writeIdx + 1) % rb.size
	rb.count++
	rb.notEmpty.Signal()
	return nil
}

func (rb *RingBuffer) Get() (interface{}, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for rb.count == 0 {
		if rb.closed {
			return nil, fmt.Errorf("buffer is closed")
		}
		rb.notEmpty.Wait()
	}

	item := rb.buffer[rb.readIdx]
	rb.buffer[rb.readIdx] = nil
	rb.readIdx = (rb.readIdx + 1) % rb.size
	rb.count--
	rb.notFull.Signal()
	return item, nil
}

func (rb *RingBuffer) Close() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.closed = true
	rb.notEmpty.Broadcast()
	rb.notFull.Broadcast()
}

func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}
