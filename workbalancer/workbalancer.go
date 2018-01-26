package workbalancer

type Workbalancer interface {
	Balance(workloads []Workload)
	GetResults() <-chan WorkResult
}

type channelBalancer struct {
	workerAvailable chan *channelWorker
	workResults     chan WorkResult
	workers         []*channelWorker
}

func NewChannelBalancer(nWorkers int) Workbalancer {
	workerAvailable := make(chan *channelWorker, nWorkers)
	workResults := make(chan WorkResult)
	workers := []*channelWorker{}
	for i := 0; i < nWorkers; i++ {
		worker := newChannelWorker(workerAvailable, workResults)
		workers = append(workers, worker)
	}
	return &channelBalancer{
		workerAvailable: workerAvailable,
		workResults:     workResults,
		workers:         workers,
	}
}

type Workload interface {
	Do() WorkResult
}

type WorkResult interface {
}

func (b *channelBalancer) Balance(workloads []Workload) {
	for _, workload := range workloads {
		worker := <-b.workerAvailable
		worker.addWork(workload)
	}
	for _, worker := range b.workers {
		worker.close()
	}
	close(b.workerAvailable)
}

func (b *channelBalancer) GetResults() <-chan WorkResult {
	return b.workResults
}
