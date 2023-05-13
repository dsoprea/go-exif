package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	log "github.com/dsoprea/go-logging"
	rifs "github.com/dsoprea/go-utility/v2/filesystem"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

func TestTagExifC4A5PrintIM_String(t *testing.T) {
	ut := TagExifA301SceneType(0x1234)

	s := ut.String()
	if s != "0x00001234" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecExifC4A5PrintIM_Encode(t *testing.T) {
	ut := TagExifC4A5PrintIM{
		version: "1234",
		values: []PrintIMKeyValue{
			{
				key:   0x1,
				value: 0x12345678,
			},
			{
				key:   0x4212,
				value: 0x90ABCDEF,
			},
		},
	}

	codec := CodecExifC4A5PrintIM{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expectedEncoded := []byte{
		0x50, 0x72, 0x69, 0x6E, 0x74, 0x49, 0x4D, 0x00, // "PrintIM" + null byte
		0x31, 0x32, 0x33, 0x34, 0x00, 0x00, 0x00, 0x02, // "1234" version, 2 null bytes, and num_chunks (big endian for tests)
		0x00, 0x01, 0x12, 0x34, 0x56, 0x78, // First chunk
		0x42, 0x12, 0x90, 0xAB, 0xCD, 0xEF, // Second chunk
	}
	if !bytes.Equal(encoded, expectedEncoded) {
		exifcommon.DumpBytesClause(encoded)
		t.Fatalf("Encoding not correct.")
	} else if unitCount != uint32(len(expectedEncoded)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodecExifC4A5PrintIM_Decode(t *testing.T) {
	expectedUt := TagExifC4A5PrintIM{
		version: "1234",
		values: []PrintIMKeyValue{
			{
				key:   0x1,
				value: 0x12345678,
			},
			{
				key:   0x4212,
				value: 0x90ABCDEF,
			},
		},
	}

	encoded := []byte{
		0x50, 0x72, 0x69, 0x6E, 0x74, 0x49, 0x4D, 0x00, // "PrintIM" + null byte
		0x31, 0x32, 0x33, 0x34, 0x00, 0x00, 0x00, 0x02, // "1234" version, 2 null bytes, and num_chunks (big endian for tests)
		0x00, 0x01, 0x12, 0x34, 0x56, 0x78, // First chunk
		0x42, 0x12, 0x90, 0xAB, 0xCD, 0xEF, // Second chunk
	}

	sb := rifs.NewSeekableBufferWithBytes(encoded)

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(encoded)),
		0,
		nil,
		sb,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := CodecExifC4A5PrintIM{}

	decoded, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(decoded, expectedUt) != true {
		t.Fatalf("Decoded struct not correct.")
	}
}
