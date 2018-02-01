package main

import (
	"time"

	"github.com/Financial-Times/publish-failure-resolver-go/workbalancer"
)

type parallelRepublisher interface {
	Republish(uuids []string, publishScope string, tidPrefix string)
}

type notifyingParallelRepublisher struct {
	republisher singleRepublisher
	balancer    workbalancer.Workbalancer
	rateLimit   time.Duration
	parallelism int
}

func newNotifyingParallelRepublisher(republisher singleRepublisher, rateLimit time.Duration, parallelism int) *notifyingParallelRepublisher {
	return &notifyingParallelRepublisher{
		republisher: republisher,
		balancer:    workbalancer.NewChannelBalancer(parallelism),
		rateLimit:   rateLimit,
	}
}

func (r *notifyingParallelRepublisher) Republish(uuids []string, publishScope string, tidPrefix string) {
	// must empty results channel
	go func() {
		for _ = range r.balancer.GetResults() {
		}
	}()
	var workloads []workbalancer.Workload
	for _, uuid := range uuids {
		workloads = append(workloads, &publishWork{
			uuid:         uuid,
			republisher:  r.republisher,
			publishScope: publishScope,
			tidPrefix:    tidPrefix,
			limiter:      time.Tick(r.rateLimit),
		})
	}
	r.balancer.Balance(workloads)
}

type publishWork struct {
	uuid         string
	publishScope string
	tidPrefix    string
	limiter      <-chan time.Time
	republisher  singleRepublisher
}

func (w *publishWork) Do() workbalancer.WorkResult {
	<-w.limiter
	w.republisher.Republish(w.uuid, w.tidPrefix, w.publishScope)
	return nil
}
