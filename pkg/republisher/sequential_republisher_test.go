package republisher

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSequentialRepublishSingle_Ok(t *testing.T) {
	mockedUUIDRepublisher := new(mockUUIDRepublisher)
	msg := OKMsg{
		uuid:                     "19cf2763-90b1-40db-90e7-e813425ebe81",
		tid:                      "prefix1",
		collectionName:           "collection1",
		collectionOriginSystemID: "originSystemId1",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	mockedUUIDRepublisher.On("Republish", "19cf2763-90b1-40db-90e7-e813425ebe81", "prefix1", ScopeBoth).Return([]*OKMsg{&msg}, []error{})

	r := NewNotifyingSequentialRepublisher(mockedUUIDRepublisher)
	r.Republish([]string{"19cf2763-90b1-40db-90e7-e813425ebe81"}, ScopeBoth, "prefix1")

	mock.AssertExpectationsForObjects(t, mockedUUIDRepublisher)
}

func TestRepublishMultiple_Ok(t *testing.T) {
	mockedUUIDRepublisher := new(mockUUIDRepublisher)
	nOk := 10
	nErr := 5
	uuids := []string{}
	for i := 0; i < nOk; i++ {
		uuids = append(uuids, "19cf2763-90b1-40db-90e7-e813425ebe81")
	}
	for i := 0; i < nErr; i++ {
		uuids = append(uuids, "70357268-04f7-4149-bb17-217d3eb56d49")
	}
	msg1 := OKMsg{
		uuid:                     "19cf2763-90b1-40db-90e7-e813425ebe81",
		tid:                      "prefix1tid_123",
		collectionName:           "collection1",
		collectionOriginSystemID: "originSystemId1",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	msg2 := OKMsg{
		uuid:                     "19cf2763-90b1-40db-90e7-e813425ebe81",
		tid:                      "prefix1tid_456",
		collectionName:           "collection2",
		collectionOriginSystemID: "originSystemId1",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	err1 := fmt.Errorf("test some error publishing 1")
	err2 := fmt.Errorf("test some error publishing 2")
	mockedUUIDRepublisher.On("Republish", "19cf2763-90b1-40db-90e7-e813425ebe81", "prefix1", ScopeBoth).Times(nOk).Return([]*OKMsg{&msg1, &msg2}, []error{})
	mockedUUIDRepublisher.On("Republish", "70357268-04f7-4149-bb17-217d3eb56d49", "prefix1", ScopeBoth).Times(nErr).Return([]*OKMsg{}, []error{err1, err2})
	r := NewNotifyingSequentialRepublisher(mockedUUIDRepublisher)
	actualMsgs, actualErrs := r.Republish(uuids, ScopeBoth, "prefix1")

	mock.AssertExpectationsForObjects(t, mockedUUIDRepublisher)
	assert.Equal(t, 2*nOk, len(actualMsgs))
	var msg1equal, msg2equal int
	for _, actualMsg := range actualMsgs {
		if msg1 == *actualMsg {
			msg1equal++
		} else if msg2 == *actualMsg {
			msg2equal++
		}
	}
	assert.Equal(t, nOk, msg1equal)
	assert.Equal(t, nOk, msg2equal)
	assert.Equal(t, 2*nErr, len(actualErrs))
	var err1equal, err2equal int
	for _, actualErr := range actualErrs {
		if err1 == actualErr {
			err1equal++
		} else if err2 == actualErr {
			err2equal++
		}
	}
	assert.Equal(t, nErr, err1equal)
	assert.Equal(t, nErr, err2equal)
}
