package workbalancer

type channelWorker struct {
	workloads       chan Workload
	workResults     chan<- WorkResult
	workerAvailable chan<- *channelWorker
}

func newChannelWorker(workerAvailable chan<- *channelWorker, workResults chan<- WorkResult) *channelWorker {
	worker := &channelWorker{
		workloads:       make(chan Workload, 1),
		workResults:     workResults,
		workerAvailable: workerAvailable,
	}
	go worker.work()
	worker.workerAvailable <- worker
	return worker
}

func (w *channelWorker) addWork(workload Workload) {
	w.workloads <- workload
}

func (w *channelWorker) work() {
	for {
		workload, more := <-w.workloads
		if !more {
			break
		}
		w.workResults <- workload.Do()
		w.workerAvailable <- w
	}
}

func (w *channelWorker) close() {
	close(w.workloads)
}
