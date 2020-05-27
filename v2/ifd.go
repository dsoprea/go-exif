package exif

import (
	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

// TODO(dustin): This file now exists for backwards-compatibility only.

// NewIfdMapping returns a new IfdMapping struct.
//
// RELEASE(dustin): This is a bridging function for backwards-compatibility. Remove this in the next release.
func NewIfdMapping() (ifdMapping *exifcommon.IfdMapping) {
	return exifcommon.NewIfdMapping()
}

// NewIfdMappingWithStandard retruns a new IfdMapping struct preloaded with the
// standard IFDs.
//
// RELEASE(dustin): This is a bridging function for backwards-compatibility. Remove this in the next release.
func NewIfdMappingWithStandard() (ifdMapping *exifcommon.IfdMapping) {
	return exifcommon.NewIfdMappingWithStandard()
}

// LoadStandardIfds loads the standard IFDs into the mapping.
//
// RELEASE(dustin): This is a bridging function for backwards-compatibility. Remove this in the next release.
func LoadStandardIfds(im *exifcommon.IfdMapping) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = exifcommon.LoadStandardIfds(im)
	log.PanicIf(err)

	return nil
}
