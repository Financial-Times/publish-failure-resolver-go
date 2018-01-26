package main

import (
	"sync"
)

type parallelRepublisher interface {
	Republish(uuids []string, republishScope string, tidPrefix string)
}

type notifyingParallelRepublisher struct {
	republishers []sequentialRepublisher
	wg           *sync.WaitGroup
}

func newNotifyingParallelRepublisher(sequentialRepublisherConstructor func() sequentialRepublisher, parallelism int) *notifyingParallelRepublisher {
	var republishers []sequentialRepublisher
	for i := 0; i < parallelism; i++ {
		republishers = append(republishers, sequentialRepublisherConstructor())
	}
	var wg sync.WaitGroup
	wg.Add(parallelism)
	return &notifyingParallelRepublisher{republishers, &wg}
}

func (r *notifyingParallelRepublisher) Republish(uuids []string, republishScope string, tidPrefix string) {
	var uuidSegments [][]string
	for i, uuid := range uuids {
		segmentI := i % len(r.republishers)
		uuidSegments[segmentI] = append(uuidSegments[segmentI], uuid)
	}

	for _, seqRepublisher := range r.republishers {
		go seqRepublisher.Republish(uuids, republishScope, tidPrefix)
	}

	r.wg.Wait()
}
