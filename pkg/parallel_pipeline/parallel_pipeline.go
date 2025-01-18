package parallelpipeline

import (
	"context"
	"sync"
)

/*
10. Parallel Pipeline:
Поддерживает произвольное количество этапов
Обеспечивает параллельную обработку на каждом этапе
Реализует backpressure
Поддерживает graceful shutdown
*/

/*
Условие:
Реализовать конвейер обработки данных с возможностью параллельного выполнения этапов.
Требования:
- Поддержка различного количества этапов
- Возможность параллельной обработки на каждом этапе
- Контроль скорости обработки (backpressure)
- Graceful shutdown всего конвейера
*/

type Stage func(ctx context.Context, in <-chan interface{}) <-chan interface{}

type Pipeline struct {
	stages  []Stage
	ctx     context.Context
	cancel  context.CancelFunc
	errChan chan error
}

func NewPipeline(stages ...Stage) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		stages:  stages,
		ctx:     ctx,
		cancel:  cancel,
		errChan: make(chan error, 1),
	}
}

func (p *Pipeline) Run(input <-chan interface{}) <-chan interface{} {
	current := input
	for _, stage := range p.stages {
		current = stage(p.ctx, current)
	}
	return current
}

func (p *Pipeline) Stop() {
	p.cancel()
}

// ParallelStage creates a stage with parallel processing capability and backpressure
func ParallelStage(workers int, bufferSize int, processor func(interface{}) (interface{}, error)) Stage {
	return func(ctx context.Context, in <-chan interface{}) <-chan interface{} {
		out := make(chan interface{}, bufferSize) // Buffered channel for backpressure
		var wg sync.WaitGroup

		// Start workers
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case item, ok := <-in:
						if !ok {
							return
						}

						result, err := processor(item)
						if err != nil {
							// Handle error, could send to error channel if needed
							continue
						}

						select {
						case <-ctx.Done():
							return
						case out <- result:
							// Successfully sent result
						}
					}
				}
			}()
		}

		// Close output channel when all workers are done
		go func() {
			wg.Wait()
			close(out)
		}()

		return out
	}
}
