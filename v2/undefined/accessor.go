package exifundefined

import (
	"reflect"

	"encoding/binary"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

func Encode(value interface{}, byteOrder binary.ByteOrder) (encoded []byte, unitCount uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	encoderName := reflect.TypeOf(value).Name()

	encoder, found := encoders[encoderName]
	if found == false {
		log.Panicf("no encoder registered for type [%s]", encoderName)
	}

	encoded, unitCount, err = encoder.Encode(value, byteOrder)
	log.PanicIf(err)

	return encoded, unitCount, nil
}

// UndefinedValue knows how to resolve the value for most unknown-type tags.
func Decode(ifdPath string, tagId uint16, valueContext *exifcommon.ValueContext, byteOrder binary.ByteOrder) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	uth := UndefinedTagHandle{
		IfdPath: ifdPath,
		TagId:   tagId,
	}

	decoder, found := decoders[uth]
	if found == false {
		// We have no choice but to return the error. We have no way of knowing how
		// much data there is without already knowing what data-type this tag is.
		return nil, exifcommon.ErrUnhandledUnknownTypedTag
	}

	if valueContext.IfdPath() != ifdPath || valueContext.TagId() != tagId {
		log.Panicf("IFD-path for codec does not match value-context: [%s] (0x%04x) != [%s] (0x%04x)", ifdPath, tagId, valueContext.IfdPath(), valueContext.TagId())
	}

	value, err = decoder.Decode(valueContext)
	log.PanicIf(err)

	return value, nil
}
