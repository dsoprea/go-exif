package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTag001CGPSAreaInformation_String(t *testing.T) {
	ut := Tag001CGPSAreaInformation{"abc"}
	s := ut.String()

	if s != "abc" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodec001CGPSAreaInformation_Encode(t *testing.T) {
	s := "abc"
	ut := Tag001CGPSAreaInformation{s}

	codec := Codec001CGPSAreaInformation{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Equal(encoded, []byte(s)) != true {
		t.Fatalf("Encoded bytes not correct: %v", encoded)
	} else if unitCount != uint32(len(s)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodec001CGPSAreaInformation_Decode(t *testing.T) {
	s := "abc"
	ut := Tag001CGPSAreaInformation{s}

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

	codec := Codec001CGPSAreaInformation{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(value, ut) != true {
		t.Fatalf("Decoded value not correct: %s\n", value)
	}
}
