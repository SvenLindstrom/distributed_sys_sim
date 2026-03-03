package worker

type WorkerManager struct {
	workers   map[string]*Worker
	ready     chan *Worker
	rpcDialer RPCDialer
}

func NewWorkerManager(queueSize int) WorkerManager {
	workers := make(chan *Worker, queueSize)
	wmap := make(map[string]*Worker)
	dialer := RealRPCDialer{}
	return WorkerManager{wmap, workers, &dialer}
}

func (w *WorkerManager) NewWorker(address string, id string) bool {
	client, err := w.rpcDialer.Dial(address)
	if err != nil {
		println(err.Error())
		return false
	}
	worker := &Worker{address, id, IDLE, "", client}
	w.workers[worker.ID] = worker
	w.ready <- worker
	return true
}

func (w *WorkerManager) JobCompleted(workerID string) {
	worker := w.workers[workerID]
	worker.JobFinished()
	w.ready <- worker
}

func (w *WorkerManager) GetWorker() *Worker {
	status := OFFLINE
	var worker *Worker
	for status == OFFLINE {
		worker = <-w.ready
		status = worker.state
	}
	return worker
}
