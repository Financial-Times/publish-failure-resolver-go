package workbalancer

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var exists = struct{}{}

func TestBalancing_Ok(t *testing.T) {
	var workloads []Workload
	n := 32
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(4)

	actualResults := make(map[int]struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for result := range balancer.GetResults() {
			intResult, ok := result.(int)
			if !ok {
				t.Fatalf("A result was not of correct type. %v", result)
			}
			actualResults[intResult] = exists
		}
		wg.Done()
	}()
	balancer.Balance(workloads)
	wg.Wait()

	for i := 0; i < n; i++ {
		_, ok := actualResults[i+1]
		assert.True(t, ok, "expected work result %d wasn't found in results", i+1)
	}
}

func TestSingleWorker_Ok(t *testing.T) {
	var workloads []Workload
	n := 8
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(1)

	actualResults := make(map[int]struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for result := range balancer.GetResults() {
			intResult, ok := result.(int)
			if !ok {
				t.Fatalf("A result was not of correct type. %v", result)
			}
			// log.Infof("result=%v", intResult)
			actualResults[intResult] = exists
		}
		wg.Done()
	}()
	balancer.Balance(workloads)
	wg.Wait()

	for i := 0; i < n; i++ {
		_, ok := actualResults[i+1]
		assert.True(t, ok, "expected work result %d wasn't found in results", i+1)
	}
}

func TestNoWorker_Ok(t *testing.T) {
	var workloads []Workload
	n := 8
	for i := 0; i < n; i++ {
		workloads = append(workloads, &incWork{i: i})
	}
	balancer := NewChannelBalancer(0)

	actualResults := make(map[int]struct{})
	wg := make(chan bool)
	go func() {
		for result := range balancer.GetResults() {
			t.Fatalf("A result appeared, but with no workers this should not have been possible. result=%v", result)
		}
		wg <- true
	}()
	timer := time.NewTimer(time.Second)
	select {
	case <-timer.C:
		return
	case <-wg:
		assert.Equal(t, actualResults, 0, "Results should be 0 because there are no workers.")
		balancer.Balance(workloads)
		t.Fatalf("Shouldn't end here as there are no workers to process work, hence no results.")

	}
}

type incWork struct {
	i int
}

func (w *incWork) Do() WorkResult {
	return w.i + 1
}
