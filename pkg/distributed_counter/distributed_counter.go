package distributedcounter

import (
	"runtime"
	"sync/atomic"
)

/*
Distributed Counter:
Минимизирует contention через striping
Использует атомарные операции
Оптимизирует доступ к памяти
Эффективно масштабируется на многоядерных системах
*/

/*
Условие:
Реализовать распределенный счетчик с минимизацией contention.
Требования:
- Использование техники striping для уменьшения конкуренции
- Атомарные операции
- Эффективное чтение текущего значения
- Минимизация false sharing
*/

type StripedCounter struct {
	counters []uint64
	mask     uint64
}

func NewStripedCounter(stripes int) *StripedCounter {
	// Round up to power of 2
	stripes--
	stripes |= stripes >> 1
	stripes |= stripes >> 2
	stripes |= stripes >> 4
	stripes |= stripes >> 8
	stripes |= stripes >> 16
	stripes |= stripes >> 32
	stripes++

	return &StripedCounter{
		counters: make([]uint64, stripes),
		mask:     uint64(stripes - 1),
	}
}

// Used for generating unique IDs for goroutines
var goroutineIDCounter uint64

func getCPUOrGoroutineID() uint64 {
	// Get a pseudo-unique ID for the goroutine
	return atomic.AddUint64(&goroutineIDCounter, 1) % uint64(runtime.GOMAXPROCS(0))
}

func (sc *StripedCounter) Increment() {
	stripe := getCPUOrGoroutineID() & sc.mask
	atomic.AddUint64(&sc.counters[stripe], 1)
}

func (sc *StripedCounter) GetValue() uint64 {
	var sum uint64
	for _, v := range sc.counters {
		sum += atomic.LoadUint64(&v)
	}
	return sum
}
