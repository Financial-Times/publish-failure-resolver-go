package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

var testCollections = map[string]targetSystem{
	"methode": {
		name:           "methode",
		originSystemID: "methode-web-pub",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
	"wordpress": {
		name:           "wordpress",
		originSystemID: "wordpress",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
	"video": {
		name:           "video",
		originSystemID: "next-video-editor",
		notifierApp:    cmsNotifier,
		scope:          scopeContent,
	},
}

func TestOkAndSoftErrors_Ok(t *testing.T) {
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedUCRepublisher := new(mockUCRepublisher)
	republisher := newNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollections)

	msg := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	var nilMsg *okMsg

	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["video"]).Return(nilMsg, false, fmt.Errorf("test error let's say 401"))

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 1, len(msgs))
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, msg, *msgs[0])
	assert.Equal(t, fmt.Errorf("error publishing test error let's say 401"), errs[0])
}

type mockDocStoreClient struct {
	mock.Mock
}

func (m *mockDocStoreClient) GetImageSetsModelUUID(setUUID, tid string) (bool, string, error) {
	args := m.Called()
	return args.Bool(0), args.String(1), args.Error(2)
}

type mockUCRepublisher struct {
	mock.Mock
}

func (m *mockUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection targetSystem) (msg *okMsg, wasFound bool, err error) {
	args := m.Called(uuid, tid, collection)
	return args.Get(0).(*okMsg), args.Bool(1), args.Error(2)
}
