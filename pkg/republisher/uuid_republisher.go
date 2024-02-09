package republisher

import (
	"fmt"
)

type UUIDRepublisher interface {
	Republish(uuid, republishScope string, tidCount int) (msgs []*OKMsg, errs []error)
}

type NotifyingUUIDRepublisher struct {
	ucRepublisher UUIDCollectionRepublisher
	collections   Collections
}

func NewNotifyingUUIDRepublisher(uuidCollectionRepublisher UUIDCollectionRepublisher, collections Collections) *NotifyingUUIDRepublisher {
	return &NotifyingUUIDRepublisher{
		ucRepublisher: uuidCollectionRepublisher,
		collections:   collections,
	}
}

func (r *NotifyingUUIDRepublisher) Republish(uuid, republishScope string, tidCount int) (msgs []*OKMsg, errs []error) {
	isFoundInAnyCollection := false
	priorityCollection := r.collections["universal-content"]
	isFoundInPriorityCollection := false

	republishFrom := func(collection CollectionMetadata) []*OKMsg {
		tid := fmt.Sprintf("tid_search_reingest_carousel_%010d_gentx", tidCount)
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

	if republishScope == ScopeBoth || republishScope == ScopeContent {
		// try priority content collection first
		msgs = republishFrom(priorityCollection)
		isFoundInPriorityCollection = isFoundInAnyCollection
		// if not found in priority, try all other content
		if !isFoundInPriorityCollection {
			for _, collection := range r.collections {
				if collection.scope == ScopeContent {
					msgs = republishFrom(collection)
				}
			}
		}
	}

	// republish metadata when scope requires it
	if republishScope == ScopeBoth || republishScope == ScopeMetadata {
		for _, collection := range r.collections {
			if collection.scope == ScopeMetadata {
				msgs = republishFrom(collection)
			}
		}
	}

	return msgs, errs
}
