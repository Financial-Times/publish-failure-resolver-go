package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

func TestRepublishOk_Ok(t *testing.T) {
	start := time.Now()
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "methode", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return([]byte("native"), true, nil)
	mockedNotifierClient := new(mockNotifierClient)
	mockedNotifierClient.On("Notify", []byte("native"), "cms-notifier", "methode-web-pub", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nil)
	republisher := newNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 500*time.Millisecond)
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
	assert.True(t, time.Now().UnixNano()-start.UnixNano() > 450000, "The time limter should restrain the call to last at least 500 milliseconds")
}

func TestRepublishNotFound_NotFound(t *testing.T) {
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "methode", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return([]byte{}, false, nil)
	mockedNotifierClient := new(mockNotifierClient)
	republisher := newNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 1*time.Millisecond)
	collection := targetSystem{
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    "cms-notifier",
		scope:          "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	assert.NoError(t, err)
	assert.False(t, wasFound)
	assert.Nil(t, msg)
}

func TestRepublishErrNative_Err(t *testing.T) {
	start := time.Now()
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "methode", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return([]byte("native"), false, fmt.Errorf("Error 401 on native client"))
	mockedNotifierClient := new(mockNotifierClient)
	republisher := newNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, time.Second)
	collection := targetSystem{
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    "cms-notifier",
		scope:          "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	assert.Error(t, fmt.Errorf("Error 401 on native client"), err)
	assert.False(t, wasFound)
	assert.Nil(t, msg)
	assert.True(t, time.Now().UnixNano()-start.UnixNano() < 950000, "The time limter should have no effect on native client.")
}

func TestRepublishErrNotifier_Err(t *testing.T) {
	start := time.Now()
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "methode", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return([]byte("native"), true, nil)
	mockedNotifierClient := new(mockNotifierClient)
	mockedNotifierClient.On("Notify", []byte("native"), "cms-notifier", "methode-web-pub", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(fmt.Errorf("error on notifier maybe 404"))
	republisher := newNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 500*time.Millisecond)
	collection := targetSystem{
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    "cms-notifier",
		scope:          "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	assert.Error(t, fmt.Errorf("error on notifier maybe 404"), err)
	assert.True(t, wasFound)
	assert.Nil(t, msg)
	assert.True(t, time.Now().UnixNano()-start.UnixNano() > 450000, "The time limter should restrain the call to last at least 500 milliseconds")
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
