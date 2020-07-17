package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTagA302CfaPattern_String(t *testing.T) {
	ut := TagA302CfaPattern{
		HorizontalRepeat: 2,
		VerticalRepeat:   3,
		CfaValue: []byte{
			0, 1, 2, 3, 4, 5,
		},
	}

	s := ut.String()

	if s != "TagA302CfaPattern<HORZ-REPEAT=(2) VERT-REPEAT=(3) CFA-VALUE=(6)>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecA302CfaPattern_Encode(t *testing.T) {
	ut := TagA302CfaPattern{
		HorizontalRepeat: 2,
		VerticalRepeat:   3,
		CfaValue: []byte{
			0, 1, 2, 3, 4,
			5, 6, 7, 8, 9,
			10, 11, 12, 13, 14,
		},
	}

	codec := CodecA302CfaPattern{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expectedBytes := []byte{
		0x00, 0x02,
		0x00, 0x03,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e,
	}

	if bytes.Equal(encoded, expectedBytes) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoded bytes not correct.")
	} else if unitCount != 19 {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodecA302CfaPattern_Decode(t *testing.T) {
	encoded := []byte{
		0x00, 0x02,
		0x00, 0x03,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05,
	}

	addressableBytes := encoded
	sb := rifs.NewSeekableBufferWithBytes(addressableBytes)

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(encoded)),
		0,
		nil,
		sb,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := CodecA302CfaPattern{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	expectedValue := TagA302CfaPattern{
		HorizontalRepeat: 2,
		VerticalRepeat:   3,
		CfaValue: []byte{
			0, 1, 2, 3, 4, 5,
		},
	}

	if reflect.DeepEqual(value, expectedValue) != true {
		t.Fatalf("Decoded value not correct: %s", value)
	}
}
