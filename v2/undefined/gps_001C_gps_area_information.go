package exifundefined

import (
	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

type Codec001CGPSAreaInformation struct {
}

func (Codec001CGPSAreaInformation) Decode(valueContext *exifcommon.ValueContext) (value interface{}, err error) {
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
		exifcommon.IfdPathStandardGps,
		0x001c,
		Codec001CGPSAreaInformation{})
}
