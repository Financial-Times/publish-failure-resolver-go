package republisher

import (
	log "github.com/sirupsen/logrus"
)

type notifyingSequentialRepublisher struct {
	uuidRepublisher UUIDRepublisher
}

func NewNotifyingSequentialRepublisher(uuidRepublisher UUIDRepublisher) BulkRepublisher {
	return &notifyingSequentialRepublisher{
		uuidRepublisher: uuidRepublisher,
	}
}

func (r *notifyingSequentialRepublisher) Republish(uuids []string, publishScope string, tidPrefix string) ([]*OKMsg, []error) {
	var msgs []*OKMsg
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
