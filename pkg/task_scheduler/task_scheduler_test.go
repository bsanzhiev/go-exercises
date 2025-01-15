package taskscheduler

import (
	"errors"
	"testing"
)

func TestTaskScheduler(t *testing.T) {
	ts := NewTaskScheduler()

	taskA := &Task{
		ID: "A",
		Execute: func() error {
			return nil
		},
	}
	taskB := &Task{
		ID:           "B",
		Dependencies: []string{"A"},
		Execute: func() error {
			return nil
		},
	}
	taskC := &Task{
		ID:           "C",
		Dependencies: []string{"B"},
		Execute: func() error {
			return nil
		},
	}

	if err := ts.AddTask(taskA); err != nil {
		t.Fatal("Failed to add task A:", err)
	}
	if err := ts.AddTask(taskB); err != nil {
		t.Fatal("Failed to add task B:", err)
	}
	if err := ts.AddTask(taskC); err != nil {
		t.Fatal("Failed to add task C:", err)
	}

	if err := ts.ExecuteAll(); err != nil {
		t.Error("ExecuteAll returned error:", err)
	}

	if taskA.Status != Completed {
		t.Error("Task A status should be Completed")
	}
	if taskB.Status != Completed {
		t.Error("Task B status should be Completed")
	}
	if taskC.Status != Completed {
		t.Error("Task C status should be Completed")
	}
}

func TestTaskSchedulerFailure(t *testing.T) {
	ts := NewTaskScheduler()

	taskA := &Task{
		ID: "FailTask",
		Execute: func() error {
			return errors.New("task error")
		},
	}
	if err := ts.AddTask(taskA); err != nil {
		t.Fatal("Failed to add FailTask:", err)
	}
	if err := ts.ExecuteAll(); err == nil {
		t.Error("Expected error for fail task, got nil")
	}
	if taskA.Status != Failed {
		t.Error("FailTask status should be Failed")
	}
}
