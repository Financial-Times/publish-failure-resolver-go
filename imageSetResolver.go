package main

import (
	"github.com/Financial-Times/uuid-utils-go"
)

type imageSetUUIDResolver interface {
	GetImageSetsModelUUID(setUUID, tid string) (found bool, modelUUID string, err error)
}

type uuidImageSetResolver struct {
}

func newUUIDImageSetResolver() imageSetUUIDResolver {
	return &uuidImageSetResolver{}
}

func (r *uuidImageSetResolver) GetImageSetsModelUUID(setUUID, tid string) (found bool, modelUUID string, err error) {
	requestedUUID, _ := uuidutils.NewUUIDFromString(setUUID)
	derivedUUID, _ := uuidutils.NewUUIDDeriverWith(uuidutils.IMAGE_SET).From(requestedUUID)
	return true, derivedUUID.String(), nil
}
