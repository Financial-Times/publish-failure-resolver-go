package main

import (
	"time"

	"github.com/Financial-Times/publish-failure-resolver-go/workbalancer"
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/sirupsen/logrus"
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
	results := r.balancer.GetResults()
	go printResults(results)
	var workloads []workbalancer.Workload
	for _, uuid := range uuids {
		workloads = append(workloads, &publishWork{
			uuid:         uuid,
			republisher:  r.republisher,
			publishScope: publishScope,
			tid:          tidPrefix + transactionidutils.NewTransactionID(),
			limiter:      time.Tick(r.rateLimit),
		})
	}
	r.balancer.Balance(workloads)
}

func printResults(results <-chan workbalancer.WorkResult) {
	for result := range results {
		pr, ok := result.(publishResult)
		if !ok {
			log.Errorf("A publish's result was not of correct type. result=%v", result)
			continue
		}
		if pr.err != nil {
			log.Errorf("Error publishing uuid=%v tid=%v: %v", pr.uuid, pr.tid, pr.err)
		} else {
			log.Infof("Sent for publish uuid=%v tid=%v", pr.uuid, pr.tid)
		}
	}
}

type publishWork struct {
	uuid         string
	publishScope string
	tid          string
	limiter      <-chan time.Time
	republisher  singleRepublisher
}

type publishResult struct {
	uuid string
	tid  string
	err  error
}

func (w *publishWork) Do() workbalancer.WorkResult {
	<-w.limiter
	err := w.republisher.Republish(w.uuid, w.tid, w.publishScope)
	return publishResult{w.uuid, w.tid, err}
}
