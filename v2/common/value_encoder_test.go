package exifcommon

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
)

func TestValueEncoder_encodeBytes__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseBytes(ed.Encoded, ed.UnitCount)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeAscii__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseAscii(ed.Encoded, ed.UnitCount)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeAsciiNoNul__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseAsciiNoNul(ed.Encoded, ed.UnitCount)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, string(expected)) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeShorts__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseShorts(ed.Encoded, ed.UnitCount, byteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeLongs__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseLongs(ed.Encoded, ed.UnitCount, byteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeRationals__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseRationals(ed.Encoded, ed.UnitCount, byteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeSignedLongs__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseSignedLongs(ed.Encoded, ed.UnitCount, byteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_encodeSignedRationals__Cycle(t *testing.T) {
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

	recovered, err := parser.ParseSignedRationals(ed.Encoded, ed.UnitCount, byteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(recovered, original) != true {
		t.Fatalf("Value not recovered correctly.")
	}
}

func TestValueEncoder_Encode__Byte(t *testing.T) {
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

func TestValueEncoder_Encode__Ascii(t *testing.T) {
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

func TestValueEncoder_Encode__Short(t *testing.T) {
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

func TestValueEncoder_Encode__Long(t *testing.T) {
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

func TestValueEncoder_Encode__Rational(t *testing.T) {
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

func TestValueEncoder_Encode__SignedLong(t *testing.T) {
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

func TestValueEncoder_Encode__SignedRational(t *testing.T) {
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

func TestValueEncoder_Encode__Timestamp(t *testing.T) {
	byteOrder := TestDefaultByteOrder
	ve := NewValueEncoder(byteOrder)

	now := time.Now()

	ed, err := ve.Encode(now)
	log.PanicIf(err)

	if ed.Type != TypeAscii {
		t.Fatalf("Timestamp not encoded as ASCII.")
	}

	expectedTimestampBytes := ExifFullTimestampString(now)

	// Leave an extra byte for the NUL.
	expected := make([]byte, len(expectedTimestampBytes)+1)
	copy(expected, expectedTimestampBytes)

	if bytes.Equal(ed.Encoded, expected) != true {
		t.Fatalf("Timestamp not encoded correctly: [%s] != [%s]", string(ed.Encoded), string(expected))
	}
}
