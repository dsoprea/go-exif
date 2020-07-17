package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTag9286UserComment_String(t *testing.T) {
	comment := "some comment"

	ut := Tag9286UserComment{
		EncodingType:  TagUndefinedType_9286_UserComment_Encoding_ASCII,
		EncodingBytes: []byte(comment),
	}

	s := ut.String()
	if s != "[ASCII] some comment" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodec9286UserComment_Encode(t *testing.T) {
	comment := "some comment"

	ut := Tag9286UserComment{
		EncodingType:  TagUndefinedType_9286_UserComment_Encoding_ASCII,
		EncodingBytes: []byte(comment),
	}

	codec := Codec9286UserComment{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	typeBytes := TagUndefinedType_9286_UserComment_Encodings[TagUndefinedType_9286_UserComment_Encoding_ASCII]
	if bytes.Equal(encoded[:8], typeBytes) != true {
		exifcommon.DumpBytesClause(encoded[:8])

		t.Fatalf("Encoding type not correct.")
	}

	if bytes.Equal(encoded[8:], []byte(comment)) != true {
		exifcommon.DumpBytesClause(encoded[8:])

		t.Fatalf("Encoded comment not correct.")
	}

	if unitCount != uint32(len(encoded)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}

	exifcommon.DumpBytesClause(encoded)

}

func TestCodec9286UserComment_Decode(t *testing.T) {
	encoded := []byte{
		0x41, 0x53, 0x43, 0x49, 0x49, 0x00, 0x00, 0x00,
		0x73, 0x6f, 0x6d, 0x65, 0x20, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
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

	codec := Codec9286UserComment{}

	decoded, err := codec.Decode(valueContext)
	log.PanicIf(err)

	comment := "some comment"

	expectedUt := Tag9286UserComment{
		EncodingType:  TagUndefinedType_9286_UserComment_Encoding_ASCII,
		EncodingBytes: []byte(comment),
	}

	if reflect.DeepEqual(decoded, expectedUt) != true {
		t.Fatalf("Decoded struct not correct.")
	}
}
