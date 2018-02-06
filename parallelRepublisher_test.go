package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestSingle_Ok(t *testing.T) {
	mockedUUIDRepublisher := new(mockUUIDRepublisher)
	okMsg := "sent for publish uuid=19cf2763-90b1-40db-90e7-e813425ebe81 tid=prefix1 collection=collection1 originSystemId=originSystemId1 size=1024 notifierApp=cms-notifier"
	mockedUUIDRepublisher.On("Republish", "19cf2763-90b1-40db-90e7-e813425ebe81", "prefix1", scopeBoth).Return([]string{okMsg}, []error{})

	pRepublisher := newNotifyingParallelRepublisher(mockedUUIDRepublisher, 1)

	pRepublisher.Republish([]string{"19cf2763-90b1-40db-90e7-e813425ebe81"}, scopeBoth, "prefix1")

	mock.AssertExpectationsForObjects(t, mockedUUIDRepublisher)
}

// func TestParallel_Ok(t *testing.T) {
// 	mockedRepublisher := new(mockedRepublisher)

// 	n := 10
// 	uuids := []string{}
// 	for i := 0; i < n; i++ {
// 		uuids = append(uuids, "19cf2763-90b1-40db-90e7-e813425ebe81")
// 	}

// 	mockedRepublisher.On("RepublishUUID", "19cf2763-90b1-40db-90e7-e813425ebe81", scopeBoth, "prefix1").Times(n).Return()

// 	pRepublisher := newNotifyingParallelRepublisher(mockedRepublisher, 2, 1)

// 	pRepublisher.Republish(uuids, scopeBoth, "prefix1")

// 	mock.AssertExpectationsForObjects(t, mockedRepublisher)
// }

type mockUUIDRepublisher struct {
	mock.Mock
}

func (m *mockUUIDRepublisher) Republish(uuid, tidPrefix string, republishScope string) (msgs []string, errs []error) {
	args := m.Called(uuid, tidPrefix, republishScope)
	return args.Get(0).([]string), args.Get(1).([]error)
}
