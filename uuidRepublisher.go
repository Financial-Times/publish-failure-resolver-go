package main

import (
	"fmt"

	"github.com/Financial-Times/transactionid-utils-go"
	log "github.com/sirupsen/logrus"
)

type uuidRepublisher interface {
	Republish(uuid, tidPrefix string, republishScope string) (msgs []*okMsg, errs []error)
}

type notifyingUUIDRepublisher struct {
	ucRepublisher    uuidCollectionRepublisher
	imageSetResolver imageSetUUIDResolver
	collections      map[string]targetSystem
}

func newNotifyingUUIDRepublisher(uuidCollectionRepublisher uuidCollectionRepublisher, imageSetResolver imageSetUUIDResolver, collections map[string]targetSystem) *notifyingUUIDRepublisher {
	return &notifyingUUIDRepublisher{
		ucRepublisher:    uuidCollectionRepublisher,
		imageSetResolver: imageSetResolver,
		collections:      collections,
	}
}

func (r *notifyingUUIDRepublisher) Republish(uuid, tidPrefix string, republishScope string) (msgs []*okMsg, errs []error) {
	isFoundInAnyCollection := false
	isScopedInAnyCollection := false

	for _, collection := range r.collections {
		if republishScope != scopeBoth && republishScope != collection.scope {
			continue
		}
		tid := tidPrefix + transactionidutils.NewTransactionID()
		isScopedInAnyCollection = true
		msg, isFound, err := r.ucRepublisher.RepublishUUIDFromCollection(uuid, tid, collection)
		if err != nil {
			errs = append(errs, fmt.Errorf("error publishing %v", err))
		}
		if isFound {
			isFoundInAnyCollection = true
		}
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}

	if !isFoundInAnyCollection && isScopedInAnyCollection {
		tid := tidPrefix + transactionidutils.NewTransactionID()
		isFoundAsImageSet, imageModelUUID, err := r.imageSetResolver.GetImageSetsModelUUID(uuid, tid)
		if err != nil {
			errs = append(errs, fmt.Errorf("couldn't check if it's an ImageSet containing an image inside because of an error uuid=%v tid=%v %v", uuid, tid, err))
			return nil, errs
		}
		if !isFoundAsImageSet {
			errs = append(errs, fmt.Errorf("can't publish uuid=%v wasn't found in any of the native-store's collections and it's not an ImageSet", uuid))
			return nil, errs
		}
		log.Infof("uuid=%v was found to be an ImageSet having an imageModelUUID=%v", uuid, imageModelUUID)
		msg, isFound, err := r.ucRepublisher.RepublishUUIDFromCollection(imageModelUUID, tid, r.collections["methode"])
		if err != nil {
			errs = append(errs, fmt.Errorf("error publishing %v", err))
			return nil, errs
		}
		if !isFound {
			errs = append(errs, fmt.Errorf("can't publish imageModelUUID=%v of imageSetUuid=%v wasn't found in native-store", imageModelUUID, uuid))
		}
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	return msgs, errs
}
