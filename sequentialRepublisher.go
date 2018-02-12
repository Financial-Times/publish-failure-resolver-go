package main

import (
	log "github.com/sirupsen/logrus"
)

type notifyingSequentialRepublisher struct {
	uuidRepublisher uuidRepublisher
}

func newNotifyingSequentialRepublisher(uuidRepublisher uuidRepublisher) bulkRepublisher {
	return &notifyingSequentialRepublisher{
		uuidRepublisher: uuidRepublisher,
	}
}

func (r *notifyingSequentialRepublisher) Republish(uuids []string, publishScope string, tidPrefix string) ([]*okMsg, []error) {
	var msgs []*okMsg
	var errs []error

	for _, uuid := range uuids {
		tmsgs, terrs := r.uuidRepublisher.Republish(uuid, tidPrefix, publishScope)

		for _, msg := range tmsgs {
			log.Info(msg)
			msgs = append(msgs, msg)
		}
		for _, err := range terrs {
			log.Error(err)
			errs = append(errs, err)
		}
	}

	return msgs, errs
}
