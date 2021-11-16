package republisher

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testCollections = Collections{
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
}

var testCollectionsSingle = Collections{
	"universal-content": {
		name:                  "universal-content",
		defaultOriginSystemID: "cct",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
}

func TestOkAndSoftErrors_Ok(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	republisher := NewNotifyingUUIDRepublisher(mockedUCRepublisher, testCollections)
	msg := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "wordpress",
		collectionOriginSystemID: "cct",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}

	msg2 := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "pac-metadata",
		collectionOriginSystemID: "http://cmdb.ft.com/systems/pac",
		sizeBytes:                1024,
		notifierAppName:          "cmsMetadataNotifier",
	}

	expectedMsgs := []*OKMsg{&msg, &msg2}
	var nilMsg *OKMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["pac-metadata"]).Return(&msg2, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(&msg, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["universal-content"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["video"]).Return(nilMsg, false, fmt.Errorf("test error let's say 401"))

	msgs, errs := republisher.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "both")

	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, 1, len(errs))
	assert.ElementsMatch(t, msgs, expectedMsgs)
	assert.Equal(t, fmt.Errorf("error publishing test error let's say 401"), errs[0])
}

func TestNotScoped_Ok(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, testCollectionsSingle)
	msg := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "wordpress",
		collectionOriginSystemID: "cct",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "b3ec9282-1073-46ad-9d44-144dad7fe956", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(&msg, true, nil)

	msgs, errs := r.Republish("b3ec9282-1073-46ad-9d44-144dad7fe956", "prefix1", "metadata")

	assert.Equal(t, 0, len(msgs))
	assert.Equal(t, 0, len(errs))
}

func TestUUIDInUCAndWordpressRepublishedInUC(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	msg := OKMsg{
		uuid:                     "64bc4319-cd22-43e9-8b12-358622d7a5ba",
		tid:                      "prefix1tid_123",
		collectionName:           "universal-content",
		collectionOriginSystemID: "cct",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}

	msg2 := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "pac-metadata",
		collectionOriginSystemID: "http://cmdb.ft.com/systems/pac",
		sizeBytes:                1024,
		notifierAppName:          "cmsMetadataNotifier",
	}

	msg3 := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "wordpress",
		collectionOriginSystemID: "cct",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}

	var nilMsg *OKMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["universal-content"]).Return(&msg, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["pac-metadata"]).Return(&msg2, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(msg3, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["video"]).Return(nilMsg, false, nil)
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, testCollections)

	msgs, errs := r.Republish("64bc4319-cd22-43e9-8b12-358622d7a5ba", "prefix1", ScopeBoth)

	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, 0, len(errs))
	for _, m := range msgs {
		switch m.collectionName {
		case msg.collectionName:
			assert.Equal(t, msg, *m)
		case msg2.collectionName:
			assert.Equal(t, msg2, *m)
		default:
			assert.Fail(t, fmt.Sprintf("%s should not be republished", m.collectionName))
		}
	}

	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["pac-metadata"]).Return(nilMsg, false, nil)
	msgs, errs = r.Republish("64bc4319-cd22-43e9-8b12-358622d7a5ba", "prefix1", ScopeContent)

	assert.Equal(t, 1, len(msgs))
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, msg, *msgs[0])
}

func TestUUIDInWordpressNotInUC(t *testing.T) {
	mockedUCRepublisher := new(mockUCRepublisher)
	r := NewNotifyingUUIDRepublisher(mockedUCRepublisher, testCollections)
	msg := OKMsg{
		uuid:                     "64bc4319-cd22-43e9-8b12-358622d7a5ba",
		tid:                      "prefix1tid_123",
		collectionName:           "wordpress",
		collectionOriginSystemID: "cct",
		sizeBytes:                1024,
		notifierAppName:          "cms-notifier",
	}
	msg2 := OKMsg{
		uuid:                     "b3ec9282-1073-46ad-9d44-144dad7fe956",
		tid:                      "prefix1",
		collectionName:           "pac-metadata",
		collectionOriginSystemID: "http://cmdb.ft.com/systems/pac",
		sizeBytes:                1024,
		notifierAppName:          "cmsMetadataNotifier",
	}

	var nilMsg *OKMsg
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["pac-metadata"]).Return(&msg2, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["universal-content"]).Return(nilMsg, false, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["wordpress"]).Return(&msg, true, nil)
	mockedUCRepublisher.On("RepublishUUIDFromCollection", "64bc4319-cd22-43e9-8b12-358622d7a5ba", mock.MatchedBy(func(tid string) bool { return strings.HasPrefix(tid, "prefix1") }), testCollections["video"]).Return(nilMsg, false, nil)

	msgs, errs := r.Republish("64bc4319-cd22-43e9-8b12-358622d7a5ba", "prefix1", ScopeBoth)

	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, 0, len(errs))
	for _, m := range msgs {
		switch m.collectionName {
		case msg.collectionName:
			assert.Equal(t, msg, *m)
		case msg2.collectionName:
			assert.Equal(t, msg2, *m)
		default:
			assert.Fail(t, fmt.Sprintf("%s should not be republished", m.collectionName))
		}
	}

	msgs, errs = r.Republish("64bc4319-cd22-43e9-8b12-358622d7a5ba", "prefix1", ScopeContent)

	assert.Equal(t, 1, len(msgs))
	assert.Equal(t, 0, len(errs))
	assert.Equal(t, msg, *msgs[0])
}

type mockUCRepublisher struct {
	mock.Mock
}

func (m *mockUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection CollectionMetadata) (msg *OKMsg, wasFound bool, err error) {
	args := m.Called(uuid, tid, collection)
	return args.Get(0).(*OKMsg), args.Bool(1), args.Error(2)
}
