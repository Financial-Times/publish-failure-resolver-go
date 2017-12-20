package main

import (
	transactionidutils "github.com/Financial-Times/transactionid-utils-go"
	log "github.com/Sirupsen/logrus"
)

type republisher interface {
	Republish(uuids []string, republishScope string, tidPrefix string)
	RepublishOne(uuid string, republishScope string, tidPrefix string)
	RepublishOneFromCollection(uuid string, republishScope string, tidPrefix string, system targetSystem) (wasFound bool)
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
	for _, system := range collections {
		isFound := r.RepublishOneFromCollection(uuid, republishScope, tidPrefix, system)
		if isFound {
			isFoundInAnyCollection = true
		}
	}
	if !isFoundInAnyCollection {
		log.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections", uuid)
	}
}

func (r *notifyingRepublisher) RepublishOneFromCollection(uuid string, republishScope string, tidPrefix string, system targetSystem) (wasFound bool) {
	if republishScope != scopeBoth && republishScope != system.scope {
		return true
	}
	nativeContent, isFound, err := r.nativeStoreClient.GetNative(system.name, uuid, "tid_test")
	if err != nil {
		log.Warnf("error while fetching native content: %v", err)
		return true
	}
	if !isFound {
		return false
	}
	tid := tidPrefix + transactionidutils.NewTransactionID()
	log.Infof("publishing uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v", uuid, tid, system.name, system.originSystemID, len(nativeContent), system.notifierApp)
	err = r.notifierClient.Notify(nativeContent, system.notifierApp, system.originSystemID, uuid, tid)
	if err != nil {
		log.Errorf("can't publish uuid=%v couldn't successfully send to notifier: %v", uuid, err)
	}
	return true
}
