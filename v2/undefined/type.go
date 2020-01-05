package exifundefined

import (
	"fmt"
	"strings"

	"crypto/sha1"
	"encoding/binary"

	"github.com/dsoprea/go-logging"

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
}

// UndefinedValueEncoder knows how to decode an undefined-type tag's value from
// bytes.
type UndefinedValueDecoder interface {
	Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error)
}

type TagUndefinedType_UnknownValue []byte

func (tutuv TagUndefinedType_UnknownValue) String() string {
	parts := make([]string, len(tutuv))
	for i, c := range tutuv {
		parts[i] = fmt.Sprintf("%02x", c)
	}

	h := sha1.New()

	_, err := h.Write(tutuv)
	log.PanicIf(err)

	digest := h.Sum(nil)

	return fmt.Sprintf("Unknown<DATA=[%s] LEN=(%d) SHA1=[%020x]>", strings.Join(parts, " "), len(tutuv), digest)
}
