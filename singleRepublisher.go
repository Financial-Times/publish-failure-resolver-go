package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type singleRepublisher interface {
	Republish(uuid, tid string, republishScope string) error
}

type notifyingSingleRepublisher struct {
	notifierClient    notifierClient
	docStoreClient    docStoreClient
	nativeStoreClient nativeStoreClientInterface
}

func newNotifyingSingleRepublisher(notifierClient notifierClient, docStoreClient docStoreClient, nativeStoreClient nativeStoreClientInterface) *notifyingSingleRepublisher {
	return &notifyingSingleRepublisher{notifierClient, docStoreClient, nativeStoreClient}
}

func (r *notifyingSingleRepublisher) Republish(uuid, tid string, republishScope string) error {
	isFoundInAnyCollection := false
	isScopedInAnyCollection := false

	for _, collection := range collections {
		if republishScope != scopeBoth && republishScope != collection.scope {
			continue
		}
		isScopedInAnyCollection = true
		isFound, err := r.republishFromCollection(uuid, tid, collection)
		if err != nil {
			return fmt.Errorf("error publishing uuid=%v collection=%v", uuid, collection)
		}
		if isFound {
			isFoundInAnyCollection = true
		}
	}

	if !isFoundInAnyCollection && isScopedInAnyCollection {
		isFoundAsImageSet, imageModelUUID, err := r.docStoreClient.GetImageSetsModelUUID(uuid, tid)
		if err != nil {
			return fmt.Errorf("couldn't get ImageModel uuid from suspected ImageSet uuid=%v tid=%v %v", uuid, tid, err)
		}
		if !isFoundAsImageSet {
			return fmt.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections and it's not an ImageSet", uuid)
		}
		log.Infof("uuid=%v was found to be an ImageSet having an imageModelUUID=%v", uuid, imageModelUUID)
		isFound, err := r.republishFromCollection(imageModelUUID, tid, collections["methode"])
		if err != nil {
			return fmt.Errorf("error publishing uuid=%v collection=methode", imageModelUUID)
		}
		if !isFound {
			log.Errorf("can't publish imageModelUUID=%v of imageSetUuid=%v wasn't found in native-store", imageModelUUID, uuid)
		}
	}
	return nil
}

func (r *notifyingSingleRepublisher) republishFromCollection(uuid, tid string, system targetSystem) (wasFound bool, err error) {
	nativeContent, isFound, err := r.nativeStoreClient.GetNative(system.name, uuid, "tid_test")
	if err != nil {
		return false, fmt.Errorf("error while fetching native content: %v", err)
	}
	if !isFound {
		return false, nil
	}

	log.Infof("publishing uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v", uuid, tid, system.name, system.originSystemID, len(nativeContent), system.notifierApp)
	err = r.notifierClient.Notify(nativeContent, system.notifierApp, system.originSystemID, uuid, tid)
	if err != nil {
		return true, fmt.Errorf("can't publish uuid=%v tid=%v couldn't successfully send to notifier: %v", uuid, tid, err)
	}
	return false, nil
}
