package republisher

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	transactionidutils "github.com/Financial-Times/transactionid-utils-go"

	"github.com/Financial-Times/publish-failure-resolver-go/pkg/image"
)

type UUIDRepublisher interface {
	Republish(uuid, tidPrefix string, republishScope string) (msgs []*OKMsg, errs []error)
}

type NotifyingUUIDRepublisher struct {
	ucRepublisher    UUIDCollectionRepublisher
	imageSetResolver image.SetUUIDResolver
	collections      Collections
}

func NewNotifyingUUIDRepublisher(uuidCollectionRepublisher UUIDCollectionRepublisher, imageSetResolver image.SetUUIDResolver, collections Collections) *NotifyingUUIDRepublisher {
	return &NotifyingUUIDRepublisher{
		ucRepublisher:    uuidCollectionRepublisher,
		imageSetResolver: imageSetResolver,
		collections:      collections,
	}
}

func (r *NotifyingUUIDRepublisher) Republish(uuid, tidPrefix string, republishScope string) (msgs []*OKMsg, errs []error) {
	isFoundInAnyCollection := false
	isScopedInAnyCollection := false
	foundInUC := false

	auxRepublish := func(collection CollectionMetadata) []*OKMsg {
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
		return msgs
	}

	// If the scope matches spark's one - independently the uuid's type, there is no way to know that at this stage-
	// 1 - it tries to republish it first in universal-content collection (spark) in case it succeed we're done
	// 2 - in case the scope is ScopeBoth (content & metadata) and it has been republished in spark then it proceed to republish in metadata
	// 3 - In case none of the above match then tries the old way, iterating all the collections, and whenever there is a match the uuid is republished. (could be more than one time)
	// this is the current behaviour.
	if republishScope == r.collections["universal-content"].scope || republishScope == ScopeBoth {
		msgs = auxRepublish(r.collections["universal-content"])
		foundInUC = isFoundInAnyCollection
	}

	if foundInUC && republishScope == ScopeBoth {
		for _, collection := range r.collections {
			if collection.scope == ScopeMetadata {
				msgs = auxRepublish(collection)
			}
		}
	} else if !isFoundInAnyCollection {
		for _, collection := range r.collections {
			if republishScope != ScopeBoth && republishScope != collection.scope {
				continue
			}
			msgs = auxRepublish(collection)
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
		msg, isFound, err := r.ucRepublisher.RepublishUUIDFromCollection(imageModelUUID, tid, r.collections["universal-content"])
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
