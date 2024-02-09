package republisher

import (
	"sync"

	"github.com/Financial-Times/workbalancer"
	log "github.com/sirupsen/logrus"
)

type BulkRepublisher interface {
	Republish(uuids []string, publishScope string, tidPrefix string) ([]*OKMsg, []error)
}

type notifyingParallelRepublisher struct {
	uuidRepublisher UUIDRepublisher
	balancer        workbalancer.WorkBalancer
}

func NewNotifyingParallelRepublisher(uuidRepublisher UUIDRepublisher, parallelism int) BulkRepublisher {
	return &notifyingParallelRepublisher{
		uuidRepublisher: uuidRepublisher,
		balancer:        workbalancer.NewChannelBalancer(parallelism),
	}
}

func (r *notifyingParallelRepublisher) Republish(uuids []string, publishScope string, tidPrefix string) ([]*OKMsg, []error) {
	var msgs []*OKMsg
	var errs []error
	allResultsFetched := sync.WaitGroup{}
	allResultsFetched.Add(1)

	go func() {
		for result := range r.balancer.GetResults() {
			pResult, ok := result.(publishResult)
			if !ok {
				log.Errorf("Work result is not of expected type: %v", result)
			}
			for _, msg := range pResult.msgs {
				log.Info(msg)
				msgs = append(msgs, msg)
			}
			for _, err := range pResult.errs {
				log.Error(err)
				errs = append(errs, err)
			}
		}
		allResultsFetched.Done()
	}()

	var workloads []workbalancer.Workload
	tidCount := 0
	for _, uuid := range uuids {
		workloads = append(workloads, &publishWork{
			uuid:            uuid,
			uuidRepublisher: r.uuidRepublisher,
			publishScope:    publishScope,
			tidCount:        tidCount,
		})

		tidCount++
	}
	r.balancer.Balance(workloads)
	allResultsFetched.Wait()
	return msgs, errs
}

type publishWork struct {
	uuid            string
	publishScope    string
	tidCount        int
	uuidRepublisher UUIDRepublisher
}

type publishResult struct {
	msgs []*OKMsg
	errs []error
}

func (w *publishWork) Do() workbalancer.Result {
	msgs, errs := w.uuidRepublisher.Republish(w.uuid, w.publishScope, w.tidCount)
	return publishResult{msgs, errs}
}
