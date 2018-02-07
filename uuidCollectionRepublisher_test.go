package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

func TestRepublish_Ok(t *testing.T) {
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "methode", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return([]byte("native"), true, nil)
	mockedNotifierClient := new(mockNotifierClient)
	mockedNotifierClient.On("Notify", []byte("native"), "cms-notifier", "methode-web-pub", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nil)
	republisher := newNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 1*time.Millisecond)
	collection := targetSystem{
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    "cms-notifier",
		scope:          "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	expectedMsg := okMsg{
		uuid:                     "f3b3b579-732b-4323-affa-a316aacad213",
		tid:                      "tid_123",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                6,
		notifierAppName:          "cms-notifier",
	}
	assert.NoError(t, err)
	assert.True(t, wasFound)
	assert.Equal(t, expectedMsg, *msg)
}

type mockNativeStoreClient struct {
	mock.Mock
}

func (m *mockNativeStoreClient) GetNative(collection, uuid, tid string) (nativeContent []byte, found bool, err error) {
	args := m.Called(collection, uuid, tid)
	return args.Get(0).([]byte), args.Bool(1), args.Error(2)
}

type mockNotifierClient struct {
	mock.Mock
}

func (m *mockNotifierClient) Notify(nativeContent []byte, notifierApp, originSystemID, uuid, tid string) error {
	args := m.Called(nativeContent, notifierApp, originSystemID, uuid, tid)
	return args.Error(0)
}
