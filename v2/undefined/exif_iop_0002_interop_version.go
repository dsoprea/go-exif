package exifundefined

import (
	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

type Codec0002InteropVersion struct {
}

func (Codec0002InteropVersion) Decode(valueContext *exifcommon.ValueContext) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext.SetUnknownValueType(exifcommon.TypeAsciiNoNul)

	valueString, err := valueContext.ReadAsciiNoNul()
	log.PanicIf(err)

	return TagUndefinedGeneralString(valueString), nil
}

func init() {
	registerDecoder(
		exifcommon.IfdPathStandardExifIop,
		0x0002,
		Codec0002InteropVersion{})
}
