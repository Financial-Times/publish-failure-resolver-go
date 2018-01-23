package main

import (
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/sirupsen/logrus"
)

type republisher interface {
	RepublishUUID(uuid string, republishScope string, tidPrefix string)
}

type notifyingRepublisher struct {
	notifierClient    notifierClient
	docStoreClient    docStoreClient
	nativeStoreClient nativeStoreClientInterface
}

func newNotifyingRepublisher(notifierClient notifierClient, docStoreClient docStoreClient, nativeStoreClient nativeStoreClientInterface) *notifyingRepublisher {
	return &notifyingRepublisher{notifierClient, docStoreClient, nativeStoreClient}
}

func (r *notifyingRepublisher) RepublishUUID(uuid string, republishScope string, tidPrefix string) {
	isFoundInAnyCollection := false
	isScopedInAnyCollection := false
	for _, system := range collections {
		if republishScope != scopeBoth && republishScope != system.scope {
			continue
		}
		isScopedInAnyCollection = true
		isFound := r.republishUUIDFromCollection(uuid, tidPrefix, system)
		if isFound {
			isFoundInAnyCollection = true
		}
	}

	if !isFoundInAnyCollection && isScopedInAnyCollection {
		serachingTid := tidPrefix + transactionidutils.NewTransactionID()
		isFoundAsImageSet, imageModelUUID, err := r.docStoreClient.GetImageSetsModelUUID(uuid, serachingTid)
		if err != nil {
			log.Errorf("couldn't get ImageModel uuid from suspected ImageSet uuid=%v searchingTid=%v %v", uuid, serachingTid, err)
		}
		if !isFoundAsImageSet {
			log.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections and it's not an ImageSet", uuid)
			return
		}
		log.Infof("uuid=%v was found to be an ImageSet having an imageModelUUID=%v", uuid, imageModelUUID)
		isFound := r.republishUUIDFromCollection(imageModelUUID, tidPrefix, collections["methode"])
		if !isFound {
			log.Errorf("can't publish imageModelUUID=%v of imageSetUuid=%v wasn't found in native-store", imageModelUUID, uuid)
		}
	}
}

func (r *notifyingRepublisher) republishUUIDFromCollection(uuid string, tidPrefix string, system targetSystem) (wasFound bool) {
	nativeContent, isFound, err := r.nativeStoreClient.GetNative(system.name, uuid, "tid_test")
	if err != nil {
		log.Warnf("error while fetching native content: %v", err)
		return false
	}
	if !isFound {
		return false
	}
	tid := tidPrefix + transactionidutils.NewTransactionID()
	log.Infof("publishing uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v", uuid, tid, system.name, system.originSystemID, len(nativeContent), system.notifierApp)
	err = r.notifierClient.Notify(nativeContent, system.notifierApp, system.originSystemID, uuid, tid)
	if err != nil {
		log.Errorf("can't publish uuid=%v tid=%v couldn't successfully send to notifier: %v", uuid, tid, err)
	}
	return true
}

//to return errors instead of logging for testability
//decide on what to expose at interface. done.
//have yourself a merry little christmas
//if not found try document store, maybe it's an image set, then try again. how? recursively or how will it work? done.
//parallelize, rate limit. done.
//scope is part of RepublishUUID not republishUUIDFromCollection? decide. done.
//fix
