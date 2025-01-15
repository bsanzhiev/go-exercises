package taskscheduler

import (
	"fmt"
	"runtime"
	"sync"
)

/*
Конкурентный планировщик задач:
Создайте планировщик задач, который может выполнять задачи
конкурентно с учетом их зависимостей.
Реализуйте механизм добавления новых задач,
отслеживания зависимостей между ними,
и параллельного выполнения независимых задач.
Используйте каналы для коммуникации между горутинами и
мьютексы для защиты общих данных.
*/

/*
Условие:
Реализовать планировщик задач с учетом их зависимостей.
Требования:
- Поддержка зависимостей между задачами
- Параллельное выполнение независимых задач
- Корректная обработка циклических зависимостей
- Отслеживание статуса выполнения задач
*/

type TaskStatus int

const (
	Pending TaskStatus = iota
	Running
	Completed
	Failed
)

type Task struct {
	ID           string
	Dependencies []string
	Execute      func() error
	Status       TaskStatus
}

type TaskScheduler struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewTaskScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks: make(map[string]*Task),
	}
}

func (ts *TaskScheduler) AddTask(task *Task) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Check for circular dependencies
	if ts.hasCircularDeps(task.ID, task.Dependencies, make(map[string]bool)) {
		return fmt.Errorf("circular dependency detected for task: %s", task.ID)
	}

	ts.tasks[task.ID] = task
	return nil
}

func (ts *TaskScheduler) hasCircularDeps(taskID string, deps []string, visited map[string]bool) bool {
	if visited[taskID] {
		return true
	}

	visited[taskID] = true
	for _, depID := range deps {
		if dep, exists := ts.tasks[depID]; exists {
			if ts.hasCircularDeps(depID, dep.Dependencies, visited) {
				return true
			}
		}
	}
	visited[taskID] = false
	return false
}

func (ts *TaskScheduler) ExecuteAll() error {
	// Create channel for task execution
	taskChan := make(chan *Task)
	errorChan := make(chan error)

	// Start worker pool
	for i := 0; i < runtime.NumCPU(); i++ {
		go ts.worker(taskChan, errorChan)
	}

	go func() {
		ts.executeReadyTasks(taskChan)
		close(taskChan)
	}()

	// Wait for all tasks to complete or error
	for range ts.tasks {
		if err := <-errorChan; err != nil {
			return err
		}
	}

	return nil
}

func (ts *TaskScheduler) worker(taskChan chan *Task, errorChan chan error) {
	for task := range taskChan {
		err := task.Execute()
		ts.mu.Lock()
		if err != nil {
			task.Status = Failed
			errorChan <- fmt.Errorf("task %s failed: %v", task.ID, err)
		} else {
			task.Status = Completed
			errorChan <- nil

			// Check if any dependent tasks now be executed
			ts.executeReadyTasks(taskChan)
		}
		ts.mu.Unlock()
	}
}

func (ts *TaskScheduler) executeReadyTasks(taskChan chan *Task) {
	for _, task := range ts.tasks {
		if task.Status == Pending && ts.canExecute(task) {
			task.Status = Running
			taskChan <- task
		}
	}
}

func (ts *TaskScheduler) canExecute(task *Task) bool {
	for _, depID := range task.Dependencies {
		if dep, exists := ts.tasks[depID]; !exists || dep.Status != Completed {
			return false
		}
	}
	return true
}
