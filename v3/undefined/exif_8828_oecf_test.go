package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTag8828Oecf_String(t *testing.T) {
	ut := Tag8828Oecf{
		Columns: 11,
		Rows:    22,
	}

	s := ut.String()

	if s != "Tag8828Oecf<COLUMNS=(11) ROWS=(22)>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodec8828Oecf_Encode(t *testing.T) {
	ut := Tag8828Oecf{
		Columns:     2,
		Rows:        22,
		ColumnNames: []string{"aa", "bb"},
		Values:      []exifcommon.SignedRational{{11, 22}},
	}

	codec := Codec8828Oecf{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expectedBytes := []byte{
		0x00, 0x02,
		0x00, 0x16,
		0x61, 0x61, 0x00, 0x62, 0x62, 0x00,
		0x00, 0x00, 0x00, 0x0b, 0x00, 0x00, 0x00, 0x16}

	if bytes.Equal(encoded, expectedBytes) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoded bytes not correct.")
	} else if unitCount != 18 {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}
}

func TestCodec8828Oecf_Decode(t *testing.T) {
	encoded := []byte{
		0x00, 0x02,
		0x00, 0x16,
		0x61, 0x61, 0x00, 0x62, 0x62, 0x00,
		0x00, 0x00, 0x00, 0x0b, 0x00, 0x00, 0x00, 0x16}

	addressableData := encoded
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(encoded)),
		0,
		nil,
		sb,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := Codec8828Oecf{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	expectedValue := Tag8828Oecf{
		Columns:     2,
		Rows:        22,
		ColumnNames: []string{"aa", "bb"},
		Values:      []exifcommon.SignedRational{{11, 22}},
	}

	if reflect.DeepEqual(value, expectedValue) != true {
		t.Fatalf("Decoded value not correct: %s", value)
	}
}
