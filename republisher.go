package main

import (
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
)

type republisher interface {
	Republish(uuids []string, republishScope string, tidPrefix string)
	RepublishOne(uuid string, republishScope string, tidPrefix string)
}

type notifyingRepublisher struct {
	notifierClient    notifierClient
	nativeStoreClient nativeStoreClientInterface
}

func newNotifyingRepublisher(notifierClient notifierClient, nativeStoreClient nativeStoreClientInterface) *notifyingRepublisher {
	return &notifyingRepublisher{notifierClient, nativeStoreClient}
}

func (r *notifyingRepublisher) Republish(uuids []string, republishScope string, tidPrefix string) {
	for _, uuid := range uuids {
		r.RepublishOne(uuid, republishScope, tidPrefix)
	}
}

func (r *notifyingRepublisher) RepublishOne(uuid string, republishScope string, tidPrefix string) {
	isFoundInAnyCollection := false
	for collectionName, collection := range collections {
		if republishScope != scopeBoth && republishScope != collection.scope {
			continue
		}
		nativeContent, isFound, err := r.nativeStoreClient.GetNative(collectionName, uuid, "tid_test")
		if err != nil {
			log.Warnf("error while fetching native content: %v", err)
			continue
		}
		if !isFound {
			continue
		}
		isFoundInAnyCollection = true
		system := collections[collectionName]
		tid := tidPrefix + transactionidutils.NewTransactionID()
		log.Infof("publishing uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v", uuid, tid, collection, system.originSystemID, len(nativeContent), system.notifierApp)
		err = r.notifierClient.Notify(nativeContent, system.notifierApp, system.originSystemID, uuid, tid)
		if err != nil {
			log.Errorf("can't publish uuid=%v couldn't successfully send to notifier: %v", uuid, err)
		}
	}
	if !isFoundInAnyCollection {
		log.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections", uuid)
	}
}
