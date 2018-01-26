package main

import "time"

type sequentialRepublisher interface {
	Republish(uuids []string, republishScope string, tidPrefix string)
}

type notifyingSequentialRepublisher struct {
	limiter     <-chan time.Time
	republisher singleRepublisher
}

func newNotifyingSequentialRepublisher(republisher singleRepublisher, rateLimit time.Duration) sequentialRepublisher {
	return &notifyingSequentialRepublisher{
		limiter:     time.Tick(rateLimit),
		republisher: republisher,
	}
}

func (r *notifyingSequentialRepublisher) Republish(uuids []string, republishScope string, tidPrefix string) {
	for _, uuid := range uuids {
		<-r.limiter
		r.republisher.Republish(uuid, republishScope, tidPrefix)
	}
	// r.wg.Done()
}
