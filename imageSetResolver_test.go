package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUuidImageSetResolver(t *testing.T) {
	r := newUUIDImageSetResolver()

	imageSetUUID := "9f365884-0c25-11e8-24ad-bec2279df517"

	found, imageModelUUID, err := r.GetImageSetsModelUUID(imageSetUUID, "tid_test")
	assert.True(t, found, "found image model UUID")
	assert.Equal(t, "9f365884-0c25-11e8-bacb-2958fde95e5e", imageModelUUID, "image model UUID")
	assert.NoError(t, err)
}
