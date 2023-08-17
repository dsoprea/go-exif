package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	log "github.com/dsoprea/go-logging"
	rifs "github.com/dsoprea/go-utility/v2/filesystem"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

func TestTagC4A5PrintImageMatching_String(t *testing.T) {
	ut := TagC4A5PrintImageMatching{
		Version: "0300",
		Value:   []byte{0x01, 0x02, 0x03, 0x04},
	}

	s := ut.String()

	if s != "TagC4A5PrintImageMatching<VERSION=(0300) BYTES=([1 2 3 4])>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecC4A5PrintImageMatching_Encode(t *testing.T) {
	rawBytes := []byte{
		0x50, 0x72, 0x69, 0x6e, 0x74, 0x49, 0x4d, 0x00,
		0x30, 0x33, 0x30, 0x30, 0x00,
		0x01, 0x02, 0x03, 0x04,
	}

	ut := TagC4A5PrintImageMatching{
		Version: "0300",
		Value:   rawBytes,
	}

	codec := CodecC4A5PrintImageMatching{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Equal(encoded, rawBytes) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoded bytes not correct.")
	} else if unitCount != 17 {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodecC4A5PrintImageMatching_Decode(t *testing.T) {
	rawBytes := []byte{
		0x50, 0x72, 0x69, 0x6e, 0x74, 0x49, 0x4d, 0x00,
		0x30, 0x33, 0x30, 0x30, 0x00,
		0x01, 0x02, 0x03, 0x04,
	}

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(rawBytes)),
		0,
		nil,
		rifs.NewSeekableBufferWithBytes(rawBytes),
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := CodecC4A5PrintImageMatching{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	expectedValue := TagC4A5PrintImageMatching{
		Version: "0300",
		Value:   rawBytes,
	}

	if reflect.DeepEqual(value, expectedValue) != true {
		t.Fatalf("Decoded value not correct: %s", value)
	}
}
