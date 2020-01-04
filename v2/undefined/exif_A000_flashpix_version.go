package exifundefined

import (
	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

type CodecA000FlashpixVersion struct {
}

func (CodecA000FlashpixVersion) Decode(valueContext *exifcommon.ValueContext) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext.SetUndefinedValueType(exifcommon.TypeAsciiNoNul)

	valueString, err := valueContext.ReadAsciiNoNul()
	log.PanicIf(err)

	return TagUndefinedGeneralString(valueString), nil
}

func init() {
	registerDecoder(
		exifcommon.IfdPathStandardExif,
		0xa000,
		CodecA000FlashpixVersion{})
}
