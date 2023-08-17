package exifundefined

import (
	"bytes"
	"encoding/binary"
	"fmt"

	log "github.com/dsoprea/go-logging"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

var PrintImageMatchingHeader = []byte{0x50, 0x72, 0x69, 0x6e, 0x74, 0x49, 0x4d, 0x00}

type TagC4A5PrintImageMatching struct {
	Version string
	Value   []byte
}

func (TagC4A5PrintImageMatching) EncoderName() string {
	return "CodecC4A5PrintImageMatching"
}

func (ev TagC4A5PrintImageMatching) String() string {
	return fmt.Sprintf("TagC4A5PrintImageMatching<VERSION=(%s) BYTES=(%v)>", ev.Version, ev.Value)
}

type CodecC4A5PrintImageMatching struct{}

func (CodecC4A5PrintImageMatching) Encode(value interface{}, byteOrder binary.ByteOrder) (encoded []byte, unitCount uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	pim, ok := value.(TagC4A5PrintImageMatching)
	if !ok {
		log.Panicf("can only encode a TagC4A5PrintImageMatching")
	}

	return pim.Value, uint32(len(pim.Value)), nil
}

func (CodecC4A5PrintImageMatching) Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error) {
	// PIM structure explanation: https://www.ozhiker.com/electronics/pjmt/jpeg_info/pim.html Looks like
	// there is no sufficient open documentation about this tag:
	// https://github.com/Exiv2/exiv2/issues/1419#issuecomment-739723214
	// For that reason we are just preserving its raw bytes
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ev := TagC4A5PrintImageMatching{}

	valueContext.SetUndefinedValueType(exifcommon.TypeByte)
	rawBytes, err := valueContext.ReadBytes()
	ev.Value = rawBytes

	if !bytes.Equal(rawBytes[0:8], PrintImageMatchingHeader[:]) {
		log.Panicf("invalid header for tag 0xC4A5 PrintImageMatching")
	}

	versionLen := bytes.IndexByte(rawBytes[8:], 0)
	ev.Version = string(rawBytes[8 : 8+versionLen])

	return ev, nil
}

func init() {
	registerEncoder(
		TagC4A5PrintImageMatching{},
		CodecC4A5PrintImageMatching{})

	registerDecoder(
		exifcommon.IfdStandardIfdIdentity.UnindexedString(),
		0xc4a5,
		CodecC4A5PrintImageMatching{})
}
