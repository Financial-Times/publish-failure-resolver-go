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

var testCollectionsSingle = map[string]targetSystem{
	"methode": {
		name:           "methode",
		originSystemID: "methode-web-pub",
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

func TestNotScoped_Ok(t *testing.T) {
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedUCRepublisher := new(mockUCRepublisher)
	republisher := newNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)
	msg := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "metadata")

	assert.Equal(t, 0, len(msgs))
	assert.Equal(t, 0, len(errs))
}

func TestFoundInNoneFoundInDocStore_Ok(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	msg := okMsg{
		uuid:                     "64bc4319-cd22-43e9-8b12-358622d7a5ba",
		tid:                      "prefix1tid_123",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	var nilMsg *okMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedDocStoreClient.On("GetImageSetsModelUUID", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") })).Return(true, "64bc4319-cd22-43e9-8b12-358622d7a5ba", nil)
	republisher := newNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 1, len(msgs))
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, msg, *msgs[0])
}

func TestFoundInNoneErrInDocStore_Err(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	msg := okMsg{
		uuid:                     "64bc4319-cd22-43e9-8b12-358622d7a5ba",
		tid:                      "prefix1tid_123",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	var nilMsg *okMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedDocStoreClient.On("GetImageSetsModelUUID", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") })).Return(false, "", fmt.Errorf("error in dsapi, maybe 401"))
	republisher := newNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 0, len(msgs))
	assert.Equal(t, 1, len(errs))
	assert.True(t, strings.HasSuffix(errs[0].Error(), "error in dsapi, maybe 401"))
}

func TestFoundInNoneAndNotFoundInDocStore_Ok(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	msg := okMsg{
		uuid:                     "64bc4319-cd22-43e9-8b12-358622d7a5ba",
		tid:                      "prefix1tid_123",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	var nilMsg *okMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedDocStoreClient.On("GetImageSetsModelUUID", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") })).Return(false, "", nil)
	republisher := newNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 0, len(msgs))
	assert.Equal(t, 1, len(errs))
	assert.True(t, strings.HasSuffix(errs[0].Error(), "wasn't found in any of the native-store's collections and it's not an ImageSet"))
}

type mockDocStoreClient struct {
	mock.Mock
}

func (m *mockDocStoreClient) GetImageSetsModelUUID(setUUID, tid string) (bool, string, error) {
	args := m.Called(setUUID, tid)
	return args.Bool(0), args.String(1), args.Error(2)
}

type mockUCRepublisher struct {
	mock.Mock
}

func (m *mockUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection targetSystem) (msg *okMsg, wasFound bool, err error) {
	args := m.Called(uuid, tid, collection)
	return args.Get(0).(*okMsg), args.Bool(1), args.Error(2)
}
