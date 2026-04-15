package worker

import (
	"dssim/internal/task"
	"testing"
)

type MockRPCClient struct {
	called      bool
	serviceName string
	args        any
	reply       any
}

func (m *MockRPCClient) Call(serviceName string, args, reply any) error {
	m.called = true
	m.serviceName = serviceName
	m.args = args
	m.reply = reply

	switch r := reply.(type) {
	case *string:
		*r = "mock-worker-id"
	case *bool:
		*r = true
	}

	return nil
}

func TestNewWorker(t *testing.T) {
	w := NewWorker("worker:9000", "scheduler:9000")

	if w.address != "worker:9000" {
		t.Errorf("Expected worker address 'worker:9000', actual '%s'", w.address)
	}

	if w.SchedulerAddr != "scheduler:9000" {
		t.Errorf("Expected scheduler address 'scheduler:9000', actual '%s'", w.SchedulerAddr)
	}

	if w.state != IDLE {
		t.Errorf(
			"Expected state IDLE, actual %v",
			w.state,
		) // check how to print actual string value
	}
}

func TestRegisterWorker(t *testing.T) {
	w := NewWorker("worker:9000", "scheduler:9000")
	mockClient := &MockRPCClient{}
	w.schedulerClient = mockClient

	err := w.registerWorker()
	if err != nil {
		t.Fatal(err)
	}

	if !mockClient.called {
		t.Errorf("Expected RPC Call to '%s'", mockClient.serviceName)
	}

	if w.ID != "mock-worker-id" {
		t.Errorf("Expected worker ID 'mock-worker-id', actual '%s'", w.ID)
	}
}

func TestCompleteTask(t *testing.T) {
	w := NewWorker("worker:9000", "scheduler:9000")
	w.ID = "mock-worker-id"
	mockClient := &MockRPCClient{}
	w.schedulerClient = mockClient

	err := w.completeTask("mock-task-id")
	if err != nil {
		t.Fatal(err)
	}

	if !mockClient.called {
		t.Errorf("Expected RPC Call to '%s'", mockClient.serviceName)
	}

	args, ok := mockClient.args.(*task.TaskResult)

	if !ok {
		t.Fatalf("Expected arg type *TaskResult, actual %T", mockClient.args)
	}

	if args.TaskID != "mock-task-id" ||
		args.WorkerID != "mock-worker-id" ||
		args.Status != "completed" {
		t.Errorf("Unexpected values in TaskResult: %+v", args)
	}

}

func TestAssignTask(t *testing.T) {
	w := NewWorker("worker:9000", "scheduler:9000")

	task := task.Task{
		ID:       "mock-task-id",
		Duration: 0,
	}
	reply := false

	err := w.AssignTask(&task, &reply)
	if err != nil {
		t.Fatal(err)
	}

	if !reply {
		t.Errorf("Expected reply to be true, actual false")
	}

	if w.currentTask != "mock-task-id" || w.state != BUSY {
		t.Errorf(
			"Expected currentTask '%s' and state BUSY, actual currentTask '%s' and state '%v'",
			task.ID,
			w.currentTask,
			w.state,
		)
	}
}

func TestExecuteTask(t *testing.T) {
	w := NewWorker("worker:9000", "scheduler:9000")
	w.ID = "mock-worker-id"
	mockClient := &MockRPCClient{}
	w.schedulerClient = mockClient

	task := task.Task{
		ID:       "mock-task-id",
		Duration: 0,
	}

	w.executeTask(task)

	if w.currentTask != "" {
		t.Errorf("Expected currentTask '', actual currentTask '%s'", w.currentTask)
	}

	if w.state != IDLE {
		t.Errorf("Expected state IDLE, actual state '%v'", w.state)
	}

	if !mockClient.called {
		t.Errorf("Expected RPC Call to '%s'", mockClient.serviceName)
	}
}
