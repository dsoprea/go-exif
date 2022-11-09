package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	exifcommon "github.com/dsoprea/go-exif/v3/common"
	log "github.com/dsoprea/go-logging"
	rifs "github.com/dsoprea/go-utility/v2/filesystem"
)

func TestTagA20CSpatialFrequencyResponse_String(t *testing.T) {
	ut := TagA20CSpatialFrequencyResponse{
		Columns:     2,
		Rows:        9,
		ColumnNames: []string{"column1", "column2"},
		Values: []exifcommon.Rational{
			{Numerator: 1, Denominator: 2},
			{Numerator: 3, Denominator: 4},
		},
	}

	s := ut.String()
	if s != "CodecA20CSpatialFrequencyResponse<COLUMNS=(2) ROWS=(9)>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecA20CSpatialFrequencyResponse_Encode(t *testing.T) {
	ut := TagA20CSpatialFrequencyResponse{
		Columns:     2,
		Rows:        9,
		ColumnNames: []string{"column1", "column2"},
		Values: []exifcommon.Rational{
			{Numerator: 1, Denominator: 2},
			{Numerator: 3, Denominator: 4},
		},
	}

	codec := CodecA20CSpatialFrequencyResponse{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expectedEncoded := []byte{
		0x00, 0x02,
		0x00, 0x09,
		0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x31, 0x00,
		0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x32, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	if bytes.Equal(encoded, expectedEncoded) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoding not correct.")
	} else if unitCount != uint32(len(encoded)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodecA20CSpatialFrequencyResponse_Decode(t *testing.T) {
	expectedUt := TagA20CSpatialFrequencyResponse{
		Columns:     2,
		Rows:        9,
		ColumnNames: []string{"column1", "column2"},
		Values: []exifcommon.Rational{
			{Numerator: 1, Denominator: 2},
			{Numerator: 3, Denominator: 4},
		},
	}

	encoded := []byte{
		0x00, 0x02,
		0x00, 0x09,
		0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x31, 0x00,
		0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x32, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
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

	codec := CodecA20CSpatialFrequencyResponse{}

	decoded, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(decoded, expectedUt) != true {
		t.Fatalf("Decoded struct not correct.")
	}
}
