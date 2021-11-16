package republisher

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Financial-Times/publish-failure-resolver-go/pkg/http/api"
)

var nm = &api.NativeMSG{
	Body:           []byte("native"),
	ContentType:    "application/json",
	OriginSystemID: "http://cmdb.ft.com/systems/cct",
}
var nmEmpty = &api.NativeMSG{
	Body:           []byte(""),
	ContentType:    "application/json",
	OriginSystemID: "http://cmdb.ft.com/systems/cct",
}

func TestRepublishOk_Ok(t *testing.T) {
	start := time.Now()
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "universal-content", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nm, true, nil)
	mockedNotifierClient := new(mockNotifierClient)
	mockedNotifierClient.On("Notify", nm, "cms-notifier", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nil)
	republisher := NewNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 500*time.Millisecond)
	collection := CollectionMetadata{
		name:                  "universal-content",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/cct",
		notifierApp:           "cms-notifier",
		scope:                 "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	expectedMsg := OKMsg{
		uuid:                     "f3b3b579-732b-4323-affa-a316aacad213",
		tid:                      "tid_123",
		collectionName:           "universal-content",
		collectionOriginSystemID: "http://cmdb.ft.com/systems/cct",
		sizeBytes:                6,
		notifierAppName:          "cms-notifier",
		contentType:              "application/json",
	}
	assert.NoError(t, err)
	assert.True(t, wasFound)
	assert.Equal(t, expectedMsg, *msg)
	assert.True(t, time.Now().UnixNano()-start.UnixNano() > 450000, "The time limter should restrain the call to last at least 500 milliseconds")
}

func TestRepublishNotFound_NotFound(t *testing.T) {
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "universal-content", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nmEmpty, false, nil)
	mockedNotifierClient := new(mockNotifierClient)
	republisher := NewNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 1*time.Millisecond)
	collection := CollectionMetadata{
		name:                  "universal-content",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/cct",
		notifierApp:           "cms-notifier",
		scope:                 "content",
	}

	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	assert.NoError(t, err)
	assert.False(t, wasFound)
	assert.Nil(t, msg)
}

func TestRepublishErrNative_Err(t *testing.T) {
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "universal-content", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nm, false, fmt.Errorf("Error 401 on native client"))
	mockedNotifierClient := new(mockNotifierClient)
	republisher := NewNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, time.Second)
	collection := CollectionMetadata{
		name:                  "universal-content",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/cct",
		notifierApp:           "cms-notifier",
		scope:                 "content",
	}

	start := time.Now()
	msg, wasFound, err := republisher.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)
	now := time.Now()

	assert.Error(t, fmt.Errorf("Error 401 on native client"), err)
	assert.False(t, wasFound)
	assert.Nil(t, msg)
	assert.WithinDuration(t, start, now, time.Nanosecond*950000, "The time limiter should have no effect on native client.")
}

func TestRepublishErrNotifier_Err(t *testing.T) {
	start := time.Now()
	mockedNativeStoreClient := new(mockNativeStoreClient)
	mockedNativeStoreClient.On("GetNative", "universal-content", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(nm, true, nil)
	mockedNotifierClient := new(mockNotifierClient)
	mockedNotifierClient.On("Notify", nm, "cms-notifier", "f3b3b579-732b-4323-affa-a316aacad213", "tid_123").Return(fmt.Errorf("error on notifier maybe 404"))
	r := NewNotifyingUCRepublisher(mockedNotifierClient, mockedNativeStoreClient, 500*time.Millisecond)
	collection := CollectionMetadata{
		name:                  "universal-content",
		defaultOriginSystemID: "cct",
		notifierApp:           "cms-notifier",
		scope:                 "content",
	}

	msg, wasFound, err := r.RepublishUUIDFromCollection("f3b3b579-732b-4323-affa-a316aacad213", "tid_123", collection)

	assert.Error(t, fmt.Errorf("error on notifier maybe 404"), err)
	assert.True(t, wasFound)
	assert.Nil(t, msg)
	assert.True(t, time.Now().UnixNano()-start.UnixNano() > 450000, "The time limter should restrain the call to last at least 500 milliseconds")
}

type mockNativeStoreClient struct {
	mock.Mock
}

func (m *mockNativeStoreClient) GetNative(collection, uuid, tid string) (nMsg *api.NativeMSG, found bool, err error) {
	args := m.Called(collection, uuid, tid)
	return args.Get(0).(*api.NativeMSG), args.Bool(1), args.Error(2)
}

type mockNotifierClient struct {
	mock.Mock
}

func (m *mockNotifierClient) Notify(nMSG *api.NativeMSG, notifierApp, uuid, tid string) error {
	args := m.Called(nMSG, notifierApp, uuid, tid)
	return args.Error(0)
}
