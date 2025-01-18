package workstealing

import (
	"sync"
)

/*
Work Stealing Pool:
Реализует эффективное распределение задач
Минимизирует простои воркеров
Использует локальные очереди для уменьшения контенции
Поддерживает динамическую балансировку нагрузки
*/

/*
Условие:
Реализовать пул горутин с возможностью кражи задач между воркерами.
Требования:
- Локальные очереди задач для каждого воркера
- Возможность кражи задач у других воркеров при простое
- Балансировка нагрузки между воркерами
- Эффективное распределение задач
*/

// Simple slice-based queue
type queue struct {
	tasks []Task
}

func (q *queue) PushBack(t Task) {
	q.tasks = append(q.tasks, t)
}

func (q *queue) PopFront() Task {
	if len(q.tasks) == 0 {
		return nil
	}
	t := q.tasks[0]
	q.tasks = q.tasks[1:]
	return t
}

func (q *queue) PopBack() Task {
	if len(q.tasks) == 0 {
		return nil
	}
	index := len(q.tasks) - 1
	t := q.tasks[index]
	q.tasks = q.tasks[:index]
	return t
}

type Task func()

type Worker struct {
	id          int
	localQueue  *queue
	stealers    []*Worker
	mu          sync.Mutex
	cond        *sync.Cond
	stolenTasks chan Task
}

type WorkStealingPool struct {
	workers []*Worker
	size    int
}

func NewWorkStealingPool(size int) *WorkStealingPool {
	pool := &WorkStealingPool{
		workers: make([]*Worker, size),
		size:    size,
	}

	// Initialize workers
	for i := 0; i < size; i++ {
		pool.workers[i] = &Worker{
			id: i,
			localQueue: &queue{
				tasks: []Task{},
			},
			stolenTasks: make(chan Task, 1),
		}
	}

	// Setup steal relationships
	for i := 0; i < size; i++ {
		worker := pool.workers[i]
		worker.stealers = make([]*Worker, size-1)
		copy(worker.stealers, pool.workers[:i])
		copy(worker.stealers[i:], pool.workers[i+1:])
		worker.cond = sync.NewCond(&worker.mu)
	}

	// Start workers
	for _, w := range pool.workers {
		go pool.runWorker(w)
	}

	return pool
}

func (pool *WorkStealingPool) Submit(task Task) {
	pool.workers[0].mu.Lock()
	defer pool.workers[0].mu.Unlock()
	pool.workers[0].localQueue.PushBack(task)
	pool.workers[0].cond.Signal()
}

func (pool *WorkStealingPool) runWorker(w *Worker) {
	for {
		task := w.getTask()
		if task == nil {
			// Try to steal from other workers
			task = w.stealTask()
		}
		if task != nil {
			task()
		}
	}
}

func (w *Worker) getTask() Task {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.localQueue.tasks) > 0 {
		return w.localQueue.PopFront()
	}
	select {
	case t := <-w.stolenTasks:
		return t
	default:
		return nil
	}
}

func (w *Worker) stealTask() Task {
	// Randomly try to steal from other workers
	for _, victim := range w.stealers {
		if task := victim.localQueue.PopBack(); task != nil {
			return task
		}
	}
	return nil
}
