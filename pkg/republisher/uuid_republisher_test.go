package republisher

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testCollections = Collections{
	"methode": {
		name:                  "methode",
		defaultOriginSystemID: "methode-web-pub",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"wordpress": {
		name:                  "wordpress",
		defaultOriginSystemID: "wordpress",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"video": {
		name:                  "video",
		defaultOriginSystemID: "next-video-editor",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"universal-content": {
		name:                  "universal-content",
		defaultOriginSystemID: "cct",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"pac-metadata": {
		name:                  "pac-metadata",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/pac",
		notifierApp:           CmsMetadataNotifier,
		scope:                 ScopeMetadata,
	},
	"v1-metadata": {
		name:                  "v1-metadata",
		defaultOriginSystemID: "methode-web-pub",
		notifierApp:           CmsMetadataNotifier,
		scope:                 ScopeMetadata,
	},
}

var testCollectionsSingle = Collections{
	"methode": {
		name:                  "methode",
		defaultOriginSystemID: "methode-web-pub",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
}

func TestOkAndSoftErrors_Ok(t *testing.T) {
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedUCRepublisher := new(mockUCRepublisher)
	republisher := NewNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollections)
	msg := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}

	msg1 := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "v1-metadata",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cmsMetadataNotifie",
	}

	msg2 := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "pac-metadata",
		collectionOriginSystemID: "http://cmdb.ft.com/systems/pac",
		sizeBytes:                1024,
		notifierAppName:          "cmsMetadataNotifie",
	}

	expectedMsgs := []*okMsg{&msg, &msg1, &msg2}
	var nilMsg *okMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["v1-metadata"]).Return(&msg1, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["pac-metadata"]).Return(&msg2, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["universal-content"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["universal-content"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["video"]).Return(nilMsg, false, fmt.Errorf("test error let's say 401"))

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 3, len(msgs))
	assert.Equal(t, 1, len(errs))
	assert.ElementsMatch(t, msgs, expectedMsgs)
	assert.Equal(t, fmt.Errorf("error publishing test error let's say 401"), errs[0])
}

func TestNotScoped_Ok(t *testing.T) {
	mockedDocStoreClient := new(mockDocStoreClient)
	mockedUCRepublisher := new(mockUCRepublisher)
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)
	msg := okMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "methode",
		collectionOriginSystemID: "methode-web-pub",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["methode"]).Return(&msg, true, nil)

	msgs, errs := r.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "metadata")

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
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := r.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

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
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := r.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

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
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, mockedDocStoreClient, testCollectionsSingle)

	msgs, errs := r.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

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

func (m *mockUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection CollectionMetadata) (msg *okMsg, wasFound bool, err error) {
	args := m.Called(uuid, tid, collection)
	return args.Get(0).(*okMsg), args.Bool(1), args.Error(2)
}
