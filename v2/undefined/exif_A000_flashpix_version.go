package exifundefined

import (
	"encoding/binary"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

type TagA000FlashpixVersion struct {
	string
}

func (TagA000FlashpixVersion) EncoderName() string {
	return "CodecA000FlashpixVersion"
}

func (fv TagA000FlashpixVersion) String() string {
	return fv.string
}

type CodecA000FlashpixVersion struct {
}

func (CodecA000FlashpixVersion) Encode(value interface{}, byteOrder binary.ByteOrder) (encoded []byte, unitCount uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	s, ok := value.(TagA000FlashpixVersion)
	if ok == false {
		log.Panicf("can only encode a TagA000FlashpixVersion")
	}

	return []byte(s.string), uint32(len(s.string)), nil
}

func (CodecA000FlashpixVersion) Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext.SetUndefinedValueType(exifcommon.TypeAsciiNoNul)

	valueString, err := valueContext.ReadAsciiNoNul()
	log.PanicIf(err)

	return TagA000FlashpixVersion{valueString}, nil
}

func init() {
	registerEncoder(
		TagA000FlashpixVersion{},
		CodecA000FlashpixVersion{})

	registerDecoder(
		exifcommon.IfdPathStandardExif,
		0xa000,
		CodecA000FlashpixVersion{})
}
