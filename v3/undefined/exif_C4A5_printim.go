package exifundefined

import (
	"fmt"

	"encoding/binary"

	log "github.com/dsoprea/go-logging"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

type PrintIMKeyValue struct {
	key   uint16
	value uint32
}
type TagExifC4A5PrintIM struct {
	version string
	values  []PrintIMKeyValue
}

func (TagExifC4A5PrintIM) EncoderName() string {
	return "CodecExifC4A5PrintIM"
}

func (af TagExifC4A5PrintIM) String() string {
	return fmt.Sprintf("Version: %s", af.version)
}

const ()

type CodecExifC4A5PrintIM struct {
}

func (CodecExifC4A5PrintIM) Encode(value interface{}, byteOrder binary.ByteOrder) (encoded []byte, unitCount uint32, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	p, ok := value.(TagExifC4A5PrintIM)
	if !ok {
		log.Panicf("can only encode a TagExifC4A5PrintIM")
	}

	size := 16 + 6*len(p.values)
	e := make([]byte, size)

	copy(e, "PrintIM")
	if len(p.version) > 4 {
		log.PanicIf("version can't be more than 4 bytes")
	}
	copy(e[8:], p.version)

	byteOrder.PutUint16(e[14:], uint16(len(p.values)))

	for i, kv := range p.values {
		byteOrder.PutUint16(e[16+6*i:], kv.key)
		byteOrder.PutUint32(e[16+6*i+2:], kv.value)
	}

	return e, uint32(size), nil
}

func (CodecExifC4A5PrintIM) Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext.SetUndefinedValueType(exifcommon.TypeByte)

	// From what I understand, this is the structure:
	// 0..7: "PrintIM" followed by a 0 byte
	// 8..11: 4 byte string version number
	// 12..13: ? (two null bytes)
	// 14..15: number of chunks, as uint16
	// 16.... num_chunks times
	//   0..1 uint16: key
	//   2..5 uint32: value

	b, err := valueContext.ReadBytes()
	log.PanicIf(err)

	if len(b) < 16 {
		log.Panic("PrintIM Tag too short")
	}

	if string(b[0:7]) != "PrintIM" || b[7] != 0 {
		log.Panic(fmt.Errorf("PrintIM Tag not starting with \"PrintIM\""))
	}

	tag := TagExifC4A5PrintIM{
		version: string(b[8:12]),
	}

	numChunks := valueContext.ByteOrder().Uint16(b[14:])
	if int(numChunks*6+16) > len(b) {
		log.Panic("Size mismatch in PrintIM Tag")
	}

	for i := 0; i < int(numChunks); i++ {
		offset := i*6 + 16
		tag.values = append(tag.values, PrintIMKeyValue{
			key:   valueContext.ByteOrder().Uint16(b[offset:]),
			value: valueContext.ByteOrder().Uint32(b[offset+2:]),
		})
	}
	return tag, nil
}

func init() {
	registerEncoder(
		TagExifC4A5PrintIM{},
		CodecExifC4A5PrintIM{})

	registerDecoder(
		exifcommon.IfdStandardIfdIdentity.UnindexedString(),
		0xc4a5,
		CodecExifC4A5PrintIM{})
}
