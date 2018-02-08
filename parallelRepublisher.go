package main

import (
	"sync"

	"github.com/Financial-Times/workbalancer"
	log "github.com/sirupsen/logrus"
)

type parallelRepublisher interface {
	Republish(uuids []string, publishScope string, tidPrefix string) ([]*okMsg, []error)
}

type notifyingParallelRepublisher struct {
	uuidRepublisher uuidRepublisher
	balancer        workbalancer.Workbalancer
	parallelism     int
}

func newNotifyingParallelRepublisher(uuidRepublisher uuidRepublisher, parallelism int) *notifyingParallelRepublisher {
	return &notifyingParallelRepublisher{
		uuidRepublisher: uuidRepublisher,
		balancer:        workbalancer.NewChannelBalancer(parallelism),
	}
}

func (r *notifyingParallelRepublisher) Republish(uuids []string, publishScope string, tidPrefix string) ([]*okMsg, []error) {
	var msgs []*okMsg
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
	for _, uuid := range uuids {
		workloads = append(workloads, &publishWork{
			uuid:            uuid,
			uuidRepublisher: r.uuidRepublisher,
			publishScope:    publishScope,
			tidPrefix:       tidPrefix,
		})
	}
	r.balancer.Balance(workloads)
	allResultsFetched.Wait()
	return msgs, errs
}

type publishWork struct {
	uuid            string
	publishScope    string
	tidPrefix       string
	uuidRepublisher uuidRepublisher
}

type publishResult struct {
	msgs []*okMsg
	errs []error
}

func (w *publishWork) Do() workbalancer.WorkResult {
	msgs, errs := w.uuidRepublisher.Republish(w.uuid, w.tidPrefix, w.publishScope)
	return publishResult{msgs, errs}
}
