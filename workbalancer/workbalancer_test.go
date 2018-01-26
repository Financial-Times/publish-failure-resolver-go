package workbalancer

import (
	"sync"
	"testing"

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

type incWork struct {
	i int
}

func (w *incWork) Do() WorkResult {
	return w.i + 1
}
