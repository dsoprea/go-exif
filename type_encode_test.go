package exif

import (
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestByteCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []byte("original text")

	ed, err := ve.encodeBytes(original)
	log.PanicIf(err)

	if ed.Type != TypeByte {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte(original)

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 13 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseBytes(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestAsciiCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := "original text"

	ed, err := ve.encodeAscii(original)
	log.PanicIf(err)

	if ed.Type != TypeAscii {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte(original)
	expected = append(expected, 0)

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 14 {
		t.Fatalf("Unit-count not correct.")
	}

	// Check that the string was recovered correctly and with the trailing NUL
	// character autostripped.

	tt := NewTagType(TypeAscii, byteOrder)
	recovered, err := tt.ParseAscii(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestAsciiNoNulCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := "original text"

	ed, err := ve.encodeAsciiNoNul(original)
	log.PanicIf(err)

	if ed.Type != TypeAsciiNoNul {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte(original)

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 13 {
		t.Fatalf("Unit-count not correct.")
	}

	// Check that the string was recovered correctly and with the trailing NUL
	// character ignored (because not expected in the context of that type).

	tt := NewTagType(TypeAsciiNoNul, byteOrder)
	recovered, err := tt.ParseAsciiNoNul(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, string(expected)) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestShortCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []uint16{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.encodeShorts(original)
	log.PanicIf(err)

	if ed.Type != TypeShort {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x11,
		0x00, 0x22,
		0x00, 0x33,
		0x00, 0x44,
		0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseShorts(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestLongCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []uint32{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.encodeLongs(original)
	log.PanicIf(err)

	if ed.Type != TypeLong {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseLongs(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestRationalCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []Rational{
		{
			Numerator:   0x11,
			Denominator: 0x22,
		},
		{
			Numerator:   0x33,
			Denominator: 0x44,
		},
		{
			Numerator:   0x55,
			Denominator: 0x66,
		},
		{
			Numerator:   0x77,
			Denominator: 0x88,
		},
		{
			Numerator:   0x99,
			Denominator: 0x00,
		},
	}

	ed, err := ve.encodeRationals(original)
	log.PanicIf(err)

	if ed.Type != TypeRational {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
		0x00, 0x00, 0x00, 0x66,
		0x00, 0x00, 0x00, 0x77,
		0x00, 0x00, 0x00, 0x88,
		0x00, 0x00, 0x00, 0x99,
		0x00, 0x00, 0x00, 0x00,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseRationals(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestSignedLongCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []int32{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.encodeSignedLongs(original)
	log.PanicIf(err)

	if ed.Type != TypeSignedLong {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseSignedLongs(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestSignedRationalCycle(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []SignedRational{
		{
			Numerator:   0x11,
			Denominator: 0x22,
		},
		{
			Numerator:   0x33,
			Denominator: 0x44,
		},
		{
			Numerator:   0x55,
			Denominator: 0x66,
		},
		{
			Numerator:   0x77,
			Denominator: 0x88,
		},
		{
			Numerator:   0x99,
			Denominator: 0x00,
		},
	}

	ed, err := ve.encodeSignedRationals(original)
	log.PanicIf(err)

	if ed.Type != TypeSignedRational {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
		0x00, 0x00, 0x00, 0x66,
		0x00, 0x00, 0x00, 0x77,
		0x00, 0x00, 0x00, 0x88,
		0x00, 0x00, 0x00, 0x99,
		0x00, 0x00, 0x00, 0x00,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}

	tt := NewTagType(ed.Type, byteOrder)
	recovered, err := tt.ParseSignedRationals(ed.Encoded, ed.UnitCount)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestEncode_Byte(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []byte("original text")

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeByte {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte(original)

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 13 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_Ascii(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := "original text"

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeAscii {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte(original)
	expected = append(expected, 0)

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 14 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_Short(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []uint16{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeShort {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x11,
		0x00, 0x22,
		0x00, 0x33,
		0x00, 0x44,
		0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_Long(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []uint32{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeLong {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_Rational(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []Rational{
		{
			Numerator:   0x11,
			Denominator: 0x22,
		},
		{
			Numerator:   0x33,
			Denominator: 0x44,
		},
		{
			Numerator:   0x55,
			Denominator: 0x66,
		},
		{
			Numerator:   0x77,
			Denominator: 0x88,
		},
		{
			Numerator:   0x99,
			Denominator: 0x00,
		},
	}

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeRational {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
		0x00, 0x00, 0x00, 0x66,
		0x00, 0x00, 0x00, 0x77,
		0x00, 0x00, 0x00, 0x88,
		0x00, 0x00, 0x00, 0x99,
		0x00, 0x00, 0x00, 0x00,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_SignedLong(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []int32{0x11, 0x22, 0x33, 0x44, 0x55}

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeSignedLong {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}
}

func TestEncode_SignedRational(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	original := []SignedRational{
		{
			Numerator:   0x11,
			Denominator: 0x22,
		},
		{
			Numerator:   0x33,
			Denominator: 0x44,
		},
		{
			Numerator:   0x55,
			Denominator: 0x66,
		},
		{
			Numerator:   0x77,
			Denominator: 0x88,
		},
		{
			Numerator:   0x99,
			Denominator: 0x00,
		},
	}

	ed, err := ve.Encode(original)
	log.PanicIf(err)

	if ed.Type != TypeSignedRational {
		t.Fatalf("IFD type not expected.")
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x11,
		0x00, 0x00, 0x00, 0x22,
		0x00, 0x00, 0x00, 0x33,
		0x00, 0x00, 0x00, 0x44,
		0x00, 0x00, 0x00, 0x55,
		0x00, 0x00, 0x00, 0x66,
		0x00, 0x00, 0x00, 0x77,
		0x00, 0x00, 0x00, 0x88,
		0x00, 0x00, 0x00, 0x99,
		0x00, 0x00, 0x00, 0x00,
	}

	if reflect.DeepEqual(ed.Encoded, expected) != true {
		t.Fatalf("Data not encoded correctly.")
	} else if ed.UnitCount != 5 {
		t.Fatalf("Unit-count not correct.")
	}
}
