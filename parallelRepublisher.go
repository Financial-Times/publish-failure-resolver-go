package main

import (
	"time"
)

type parallelRepublisher interface {
	Republish(uuids []string, republishScope string, tidPrefix string)
}

type notifyingParallelRepublisher struct {
	republisher republisher
	queues      []*queueWithLimiter
}

type queueWithLimiter struct {
	l <-chan time.Time
	q chan string
}

func newNotifyingParallelRepublisher(republisher republisher, parallelism int, rateLimit time.Duration) *notifyingParallelRepublisher {
	var queues []*queueWithLimiter
	for i := 0; i < parallelism; i++ {
		l := time.Tick(rateLimit)
		q := make(chan string, 1)
		queues = append(queues, &queueWithLimiter{l, q})
	}
	return &notifyingParallelRepublisher{republisher, queues}
}

func (r *notifyingParallelRepublisher) Republish(uuids []string, republishScope string, tidPrefix string) {
	for _, ql := range r.queues {
		go r.republishFromQueue(ql, republishScope, tidPrefix)
	}

	for i, uuid := range uuids {
		qi := i % len(r.queues)
		r.queues[qi].q <- uuid
	}

	for _, ql := range r.queues {
		close(ql.q)
	}
}

func (r *notifyingParallelRepublisher) republishFromQueue(ql *queueWithLimiter, republishScope string, tidPrefix string) {
	for uuid := range ql.q {
		<-ql.l
		r.republisher.RepublishUUID(uuid, republishScope, tidPrefix)
	}
}
