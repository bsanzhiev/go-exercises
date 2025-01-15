package parallelmap

import "sync"

/*
Параллельный map-reduce:
Реализуйте функцию параллельного выполнения операций map и reduce
на большом наборе данных.
Используйте горутины для параллельной обработки данных,
каналы для сбора промежуточных результатов,
и sync.WaitGroup для синхронизации завершения всех операций.
*/

/*
Условие:
Реализовать параллельное выполнение операций map и reduce на большом наборе данных.
Требования:
- Параллельная обработка данных
- Сбор промежуточных результатов через каналы
- Использование sync.WaitGroup для синхронизации
- Возможность задать количество воркеров
*/

type MapReduceJob struct {
	Data     []interface{}
	MapFn    func(interface{}) interface{}
	ReduceFn func([]interface{}) interface{}
	Workers  int
}

func ParellelMapReduce(job MapReduceJob) interface{} {
	var wg sync.WaitGroup
	resultChan := make(chan interface{}, len(job.Data))

	// Split data into chanks
	chankSize := (len(job.Data) + job.Workers - 1) / job.Workers

	// Map phase
	for i := 0; i < len(job.Data); i += chankSize {
		end := i + chankSize
		if end > len(job.Data) {
			end = len(job.Data)
		}

		wg.Add(1)
		go func(chank []interface{}) {
			defer wg.Done()
			for _, item := range chank {
				resultChan <- job.MapFn(item)
			}
		}(job.Data[i:end])
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var mappedResult []interface{}
	for result := range resultChan {
		mappedResult = append(mappedResult, result)
	}

	return job.ReduceFn(mappedResult)
}
