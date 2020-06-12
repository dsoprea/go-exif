package exif

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestTagType_EncodeDecode_Byte(t *testing.T) {
	tt := NewTagType(TypeByte, TestDefaultByteOrder)

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, data) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseBytes(encoded, uint32(len(data)))
	log.PanicIf(err)

	if bytes.Compare(restored, data) != 0 {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_Ascii(t *testing.T) {
	tt := NewTagType(TypeAscii, TestDefaultByteOrder)

	data := "hello"

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if string(encoded) != fmt.Sprintf("%s\000", data) {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseAscii(encoded, uint32(len(data)))
	log.PanicIf(err)

	if restored != data {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_Shorts(t *testing.T) {
	tt := NewTagType(TypeShort, TestDefaultByteOrder)

	data := []uint16{0x11, 0x22, 0x33}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, []byte{0x00, 0x11, 0x00, 0x22, 0x00, 0x33}) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseShorts(encoded, uint32(len(data)))
	log.PanicIf(err)

	if reflect.DeepEqual(restored, data) != true {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_Long(t *testing.T) {
	tt := NewTagType(TypeLong, TestDefaultByteOrder)

	data := []uint32{0x11, 0x22, 0x33}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, []byte{0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33}) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseLongs(encoded, uint32(len(data)))
	log.PanicIf(err)

	if reflect.DeepEqual(restored, data) != true {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_Rational(t *testing.T) {
	tt := NewTagType(TypeRational, TestDefaultByteOrder)

	data := []Rational{
		{Numerator: 0x11, Denominator: 0x22},
		{Numerator: 0x33, Denominator: 0x44},
	}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, []byte{0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0x00, 0x44}) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseRationals(encoded, uint32(len(data)))
	log.PanicIf(err)

	if reflect.DeepEqual(restored, data) != true {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_SignedLong(t *testing.T) {
	tt := NewTagType(TypeSignedLong, TestDefaultByteOrder)

	data := []int32{0x11, 0x22, 0x33}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, []byte{0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33}) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseSignedLongs(encoded, uint32(len(data)))
	log.PanicIf(err)

	if reflect.DeepEqual(restored, data) != true {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_SignedRational(t *testing.T) {
	tt := NewTagType(TypeSignedRational, TestDefaultByteOrder)

	data := []SignedRational{
		{Numerator: 0x11, Denominator: 0x22},
		{Numerator: 0x33, Denominator: 0x44},
	}

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if bytes.Compare(encoded, []byte{0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0x00, 0x44}) != 0 {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseSignedRationals(encoded, uint32(len(data)))
	log.PanicIf(err)

	if reflect.DeepEqual(restored, data) != true {
		t.Fatalf("Data not decoded correctly.")
	}
}

func TestTagType_EncodeDecode_AsciiNoNul(t *testing.T) {
	tt := NewTagType(TypeAsciiNoNul, TestDefaultByteOrder)

	data := "hello"

	encoded, err := tt.Encode(data)
	log.PanicIf(err)

	if string(encoded) != data {
		t.Fatalf("Data not encoded correctly.")
	}

	restored, err := tt.ParseAsciiNoNul(encoded, uint32(len(data)))
	log.PanicIf(err)

	if restored != data {
		t.Fatalf("Data not decoded correctly.")
	}
}

// TODO(dustin): Add tests for TypeUndefined.

func TestTagType_FromString_Undefined(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")

			log.Panic(err)
		}
	}()

	tt := NewTagType(TypeUndefined, TestDefaultByteOrder)

	_, err := tt.FromString("")
	if err == nil {
		t.Fatalf("no error for undefined-type")
	} else if err.Error() != "undefined-type values are not supported" {
		fmt.Printf("[%s]\n", err.Error())
		log.Panic(err)
	}
}

func TestTagType_FromString_Byte(t *testing.T) {
	tt := NewTagType(TypeByte, TestDefaultByteOrder)

	value, err := tt.FromString("abc")
	log.PanicIf(err)

	if reflect.DeepEqual(value, []byte{'a', 'b', 'c'}) != true {
		t.Fatalf("byte value not correct")
	}
}

func TestTagType_FromString_Ascii(t *testing.T) {
	tt := NewTagType(TypeAscii, TestDefaultByteOrder)

	value, err := tt.FromString("abc")
	log.PanicIf(err)

	if reflect.DeepEqual(value, "abc") != true {
		t.Fatalf("ASCII value not correct: [%s]", value)
	}
}

func TestTagType_FromString_Short(t *testing.T) {
	tt := NewTagType(TypeShort, TestDefaultByteOrder)

	value, err := tt.FromString("55")
	log.PanicIf(err)

	if reflect.DeepEqual(value, uint16(55)) != true {
		t.Fatalf("short value not correct")
	}
}

func TestTagType_FromString_Long(t *testing.T) {
	tt := NewTagType(TypeLong, TestDefaultByteOrder)

	value, err := tt.FromString("66000")
	log.PanicIf(err)

	if reflect.DeepEqual(value, uint32(66000)) != true {
		t.Fatalf("long value not correct")
	}
}

func TestTagType_FromString_Rational(t *testing.T) {
	tt := NewTagType(TypeRational, TestDefaultByteOrder)

	value, err := tt.FromString("12/34")
	log.PanicIf(err)

	expected := Rational{
		Numerator:   12,
		Denominator: 34,
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("rational value not correct")
	}
}

func TestTagType_FromString_SignedLong(t *testing.T) {
	tt := NewTagType(TypeSignedLong, TestDefaultByteOrder)

	value, err := tt.FromString("-66000")
	log.PanicIf(err)

	if reflect.DeepEqual(value, int32(-66000)) != true {
		t.Fatalf("signed-long value not correct")
	}
}

func TestTagType_FromString_SignedRational(t *testing.T) {
	tt := NewTagType(TypeSignedRational, TestDefaultByteOrder)

	value, err := tt.FromString("-12/34")
	log.PanicIf(err)

	expected := SignedRational{
		Numerator:   -12,
		Denominator: 34,
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("signd-rational value not correct")
	}
}

func TestTagType_FromString_AsciiNoNul(t *testing.T) {
	tt := NewTagType(TypeAsciiNoNul, TestDefaultByteOrder)

	value, err := tt.FromString("abc")
	log.PanicIf(err)

	if reflect.DeepEqual(value, "abc") != true {
		t.Fatalf("ASCII-no-nul value not correct")
	}
}
