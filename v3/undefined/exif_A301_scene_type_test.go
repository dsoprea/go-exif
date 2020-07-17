package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTagExifA301SceneType_String(t *testing.T) {
	ut := TagExifA301SceneType(0x1234)

	s := ut.String()
	if s != "0x00001234" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecExifA301SceneType_Encode(t *testing.T) {
	ut := TagExifA301SceneType(0x1234)

	codec := CodecExifA301SceneType{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expectedEncoded := []byte{0, 0, 0x12, 0x34}

	if bytes.Equal(encoded, expectedEncoded) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoding not correct.")
	} else if unitCount != 1 {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodecExifA301SceneType_Decode(t *testing.T) {
	expectedUt := TagExifA301SceneType(0x1234)

	encoded := []byte{0, 0, 0x12, 0x34}

	rawValueOffset := encoded

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		1,
		0,
		rawValueOffset,
		nil,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := CodecExifA301SceneType{}

	decoded, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(decoded, expectedUt) != true {
		t.Fatalf("Decoded struct not correct.")
	}
}
