package main

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestSingle_Ok(t *testing.T) {
	mockedRepublisher := new(mockedRepublisher)

	mockedRepublisher.On("RepublishUUID", "19cf2763-90b1-40db-90e7-e813425ebe81", scopeBoth, "prefix1").Return()

	pRepublisher := newNotifyingParallelRepublisher(mockedRepublisher, 1, 1)

	pRepublisher.Republish([]string{"19cf2763-90b1-40db-90e7-e813425ebe81"}, scopeBoth, "prefix1")

	mock.AssertExpectationsForObjects(t, mockedRepublisher)
}

func TestParallel_Ok(t *testing.T) {
	mockedRepublisher := new(mockedRepublisher)

	n := 10
	uuids := []string{}
	for i := 0; i < n; i++ {
		uuids = append(uuids, "19cf2763-90b1-40db-90e7-e813425ebe81")
	}

	mockedRepublisher.On("RepublishUUID", "19cf2763-90b1-40db-90e7-e813425ebe81", scopeBoth, "prefix1").Times(n).Return()

	pRepublisher := newNotifyingParallelRepublisher(mockedRepublisher, 2, 1)

	pRepublisher.Republish(uuids, scopeBoth, "prefix1")

	mock.AssertExpectationsForObjects(t, mockedRepublisher)
}

type mockedRepublisher struct {
	mock.Mock
}

func (m *mockedRepublisher) RepublishUUID(uuid string, republishScope string, tidPrefix string) {
	_ = m.Called(uuid, republishScope, tidPrefix)
}
