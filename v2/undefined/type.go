package exifundefined

import (
	"encoding/binary"

	"github.com/dsoprea/go-exif/v2/common"
)

const (
	UnparseableUnknownTagValuePlaceholder = "!UNKNOWN"
)

// UndefinedValueEncoder knows how to encode an undefined-type tag's value to
// bytes.
type UndefinedValueEncoder interface {
	Encode(value interface{}, byteOrder binary.ByteOrder) (encoded []byte, unitCount uint32, err error)
}

type EncodeableValue interface {
	EncoderName() string
	String() string
}

// UndefinedValueEncoder knows how to decode an undefined-type tag's value from
// bytes.
type UndefinedValueDecoder interface {
	Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error)
}
