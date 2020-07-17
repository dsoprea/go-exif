package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTag927CMakerNote_String(t *testing.T) {
	ut := Tag927CMakerNote{
		MakerNoteType:  []byte{0, 1, 2, 3, 4},
		MakerNoteBytes: []byte{5, 6, 7, 8, 9},
	}

	s := ut.String()
	if s != "MakerNote<TYPE-ID=[00 01 02 03 04] LEN=(5) SHA1=[bdb42cb7eb76e64efe49b22369b404c67b0af55a]>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodec927CMakerNote_Encode(t *testing.T) {
	codec := Codec927CMakerNote{}

	prefix := []byte{0, 1, 2, 3, 4}
	b := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	ut := Tag927CMakerNote{
		MakerNoteType:  prefix,
		MakerNoteBytes: b,
	}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Equal(encoded, b) != true {
		t.Fatalf("Encoding not correct: %v", encoded)
	} else if unitCount != uint32(len(b)) {
		t.Fatalf("Unit-count not correct: (%d)", len(b))
	}
}

func TestCodec927CMakerNote_Decode(t *testing.T) {
	b := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	ut := Tag927CMakerNote{
		MakerNoteType:  b,
		MakerNoteBytes: b,
	}

	sb := rifs.NewSeekableBufferWithBytes(b)

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(b)),
		0,
		nil,
		sb,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := Codec927CMakerNote{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(value, ut) != true {
		t.Fatalf("Decoded value not correct: %s != %s", value, ut)
	}
}
