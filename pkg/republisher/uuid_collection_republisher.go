package republisher

import (
	"fmt"
	"github.com/Financial-Times/publish-failure-resolver-go/pkg/http/api"
	"time"
)

type UUIDCollectionRepublisher interface {
	RepublishUUIDFromCollection(uuid, tid string, collection CollectionMetadata) (msg *okMsg, wasFound bool, err error)
}

type notifyingUCRepublisher struct {
	notifierClient    api.NotifierClient
	nativeStoreClient api.NativeStoreClientInterface
	rateLimit         time.Duration
}

func NewNotifyingUCRepublisher(notifierClient api.NotifierClient, nativeStoreClient api.NativeStoreClientInterface, rateLimit time.Duration) *notifyingUCRepublisher {
	return &notifyingUCRepublisher{notifierClient, nativeStoreClient, rateLimit}
}

type okMsg struct {
	uuid                     string
	tid                      string
	collectionName           string
	collectionOriginSystemID string
	sizeBytes                int
	notifierAppName          string
	contentType              string
}

func (msg *okMsg) String() string {
	return fmt.Sprintf("sent for publish uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v contentType=%v", msg.uuid, msg.tid, msg.collectionName, msg.collectionOriginSystemID, msg.sizeBytes, msg.notifierAppName, msg.contentType)
}

func (r *notifyingUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection CollectionMetadata) (msg *okMsg, wasFound bool, err error) {
	start := time.Now()
	nativeContent, isFound, err := r.nativeStoreClient.GetNative(collection.name, uuid, tid)
	if err != nil {
		return nil, false, fmt.Errorf("error while fetching native content: %v", err)
	}
	if !isFound {
		return nil, false, nil
	}
	if nativeContent.OriginSystemID == "" {
		nativeContent.OriginSystemID = collection.defaultOriginSystemID
	}
	err = r.notifierClient.Notify(nativeContent, collection.notifierApp, uuid, tid)
	if err != nil {
		extendTimeToLength(start, r.rateLimit)
		return nil, true, fmt.Errorf("couldn't send to notifier uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v %v", uuid, tid, collection.name, collection.defaultOriginSystemID, len(nativeContent.Body), collection.notifierApp, err)
	}

	extendTimeToLength(start, r.rateLimit)
	return &okMsg{
		uuid:                     uuid,
		tid:                      tid,
		collectionName:           collection.name,
		collectionOriginSystemID: nativeContent.OriginSystemID,
		sizeBytes:                len(nativeContent.Body),
		notifierAppName:          collection.notifierApp,
		contentType:              nativeContent.ContentType,
	}, true, nil
}

func extendTimeToLength(start time.Time, length time.Duration) {
	time.Sleep(time.Duration(start.Add(length).UnixNano()-time.Now().UnixNano()) * time.Nanosecond)
}
