package events

/*
Система обработки событий с приоритетами:
Разработайте систему обработки событий, где события имеют разные приоритеты.
Используйте несколько горутин-обработчиков и каналы для распределения событий.
Реализуйте механизм, гарантирующий, что события с высоким приоритетом
обрабатываются быстрее, чем с низким.
*/

/*
Условие:
Реализовать систему обработки событий с разными приоритетами.
Требования:
- События имеют приоритеты (высокий, средний, низкий)
- Несколько горутин-обработчиков
- События с высоким приоритетом должны обрабатываться быстрее
- Корректная обработка завершения работы системы
*/

import (
	"fmt"
	"sync"
)

type Priority int

const (
	HighPriority Priority = iota
	NormalPriority
	LowPriority
)

type Event struct {
	ID       int
	Priority Priority
	Data     interface{}
}

type EventProcessor struct {
	highPriorityChan   chan Event
	normalPriorityChan chan Event
	lowPriorityChan    chan Event
	done               chan struct{}
	wg                 sync.WaitGroup
}

func NewEventProcessor(numWorker int) *EventProcessor {
	ep := &EventProcessor{
		highPriorityChan:   make(chan Event, 100),
		normalPriorityChan: make(chan Event, 100),
		lowPriorityChan:    make(chan Event, 100),
		done:               make(chan struct{}),
	}

	for i := 0; i < numWorker; i++ {
		ep.wg.Add(1)
		go ep.worker()
	}

	return ep
}

func (ep *EventProcessor) worker() {
	defer ep.wg.Done()

	for {
		select {
		case event := <-ep.highPriorityChan:
			ep.processEvent(event)
		case <-ep.done:
			return
		default:
			select {
			case event := <-ep.highPriorityChan:
				ep.processEvent(event)
			case event := <-ep.normalPriorityChan:
				ep.processEvent(event)
			case event := <-ep.lowPriorityChan:
				ep.processEvent(event)
			case <-ep.done:
				return
			}
		}
	}
}

func (ep *EventProcessor) processEvent(event Event) {
	fmt.Printf("Processing event %d with priority %d\n", event.ID, event.Priority)
}

func (ep *EventProcessor) Submit(event Event) {
	switch event.Priority {
	case HighPriority:
		ep.highPriorityChan <- event
	case NormalPriority:
		ep.normalPriorityChan <- event
	case LowPriority:
		ep.lowPriorityChan <- event
	}
}

func (ep *EventProcessor) Stop() {
	close(ep.done)
	ep.wg.Wait()
}
