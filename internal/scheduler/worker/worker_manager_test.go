package worker

import (
	"dssim/internal/job"
	"dssim/internal/network"
	"testing"
)

type MockRPCClient struct{}

func (mr *MockRPCClient) Call(serviceName string, args any, reply any) error {
	ptr := reply.(*bool)
	*ptr = true
	return nil
}

type MockDialer struct{}

func (md *MockDialer) Dial(address string) (network.RPCClient, error) {
	return &MockRPCClient{}, nil
}

func newTestWorkerManager() WorkerManagerImple {
	tmap := make(map[string]*Worker)
	tchan := make(chan *Worker, 2)

	return WorkerManagerImple{tmap, tchan, &MockDialer{}}
}

func TestNewWorker(t *testing.T) {
	manager := newTestWorkerManager()
	ok := manager.NewWorker("ip Address", "testID")
	if !ok {
		t.Fatal("expected success")
	}
	worker := manager.workers["testID"]
	if worker.IP != "ip Address" {
		t.Fatal("ip not set correctly")
	}
	if worker.ID != "testID" {
		t.Fatal("id not set correctly")
	}
	if worker.job != "" {
		t.Fatal("job not set correctly")
	}
	if worker.state != IDLE {
		t.Fatal("state not set correctly")
	}
}

func TestGetWorker(t *testing.T) {
	manager := newTestWorkerManager()
	ok := manager.NewWorker("ip Address", "testID")
	if !ok {
		t.Fatal("expected success")
	}
	worker := manager.GetWorker()
	if worker.IP != "ip Address" {
		t.Fatal(worker)
	}
	if worker.ID != "testID" {
		t.Fatal("id not set correctly")
	}
}

func TestJobCompleted(t *testing.T) {
	manager := newTestWorkerManager()
	ok := manager.NewWorker("ip Address", "testID")
	if !ok {
		t.Fatal("expected success")
	}
	worker := manager.GetWorker()

	job := &job.Job{ID: "jobID", Duration: 123}

	worker.AssignJob(job)

	manager.JobCompleted("testID")

	if worker.job != "" {
		t.Fatal("job id in worker not reset")
	}
	if worker.state != READY {
		t.Fatal(worker)
		t.Fatal("worker state not reset")
	}

	nextWorker := manager.GetWorker()
	if nextWorker.ID != "testID" {
		t.Fatal("worker not placed back in channel")
	}
}
