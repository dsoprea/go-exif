package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTag9000ExifVersion_String(t *testing.T) {
	ut := Tag9000ExifVersion{"abc"}
	s := ut.String()

	if s != "abc" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodec9000ExifVersion_Encode(t *testing.T) {
	s := "abc"
	ut := Tag9000ExifVersion{s}

	codec := Codec9000ExifVersion{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Equal(encoded, []byte(s)) != true {
		t.Fatalf("Encoded bytes not correct: %v", encoded)
	} else if unitCount != uint32(len(s)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodec9000ExifVersion_Decode(t *testing.T) {
	s := "abc"
	ut := Tag9000ExifVersion{s}

	encoded := []byte(s)

	rawValueOffset := encoded

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(encoded)),
		0,
		rawValueOffset,
		nil,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := Codec9000ExifVersion{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(value, ut) != true {
		t.Fatalf("Decoded value not correct: %s\n", value)
	}
}
