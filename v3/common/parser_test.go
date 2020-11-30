package exifcommon

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestParser_ParseBytes(t *testing.T) {
	p := new(Parser)

	encoded := []byte("abcdefg")

	value, err := p.ParseBytes(encoded, 1)
	log.PanicIf(err)

	if bytes.Equal(value, encoded[:1]) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	value, err = p.ParseBytes(encoded, 4)
	log.PanicIf(err)

	if bytes.Equal(value, encoded[:4]) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}

	value, err = p.ParseBytes(encoded, uint32(len(encoded)))
	log.PanicIf(err)

	if bytes.Equal(value, encoded) != true {
		t.Fatalf("Encoding not correct (3): %v", value)
	}
}

func TestParser_ParseAscii(t *testing.T) {
	p := new(Parser)

	original := "abcdefg"

	encoded := []byte(original[:1])
	encoded = append(encoded, 0)

	value, err := p.ParseAscii(encoded, uint32(len(encoded)))
	log.PanicIf(err)

	if value != original[:1] {
		t.Fatalf("Encoding not correct (1): %s", value)
	}

	encoded = []byte(original[:4])
	encoded = append(encoded, 0)

	value, err = p.ParseAscii(encoded, uint32(len(encoded)))
	log.PanicIf(err)

	if value != original[:4] {
		t.Fatalf("Encoding not correct (2): %v", value)
	}

	encoded = []byte(original)
	encoded = append(encoded, 0)

	value, err = p.ParseAscii(encoded, uint32(len(encoded)))
	log.PanicIf(err)

	if value != original {
		t.Fatalf("Encoding not correct (3): %v", value)
	}
}

func TestParser_ParseAsciiNoNul(t *testing.T) {
	p := new(Parser)

	original := "abcdefg"

	encoded := []byte(original)

	value, err := p.ParseAsciiNoNul(encoded, 1)
	log.PanicIf(err)

	if value != original[:1] {
		t.Fatalf("Encoding not correct (1): %s", value)
	}

	value, err = p.ParseAsciiNoNul(encoded, 4)
	log.PanicIf(err)

	if value != original[:4] {
		t.Fatalf("Encoding not correct (2): %v", value)
	}

	value, err = p.ParseAsciiNoNul(encoded, uint32(len(encoded)))
	log.PanicIf(err)

	if value != original {
		t.Fatalf("Encoding not correct (3): (%d) %v", len(value), value)
	}
}

func TestParser_ParseShorts__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x01}

	value, err := p.ParseShorts(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint16{1}) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	encoded = []byte{0x00, 0x01, 0x00, 0x02}

	value, err = p.ParseShorts(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint16{1}) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseShorts__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x01, 0x00, 0x02}

	value, err := p.ParseShorts(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint16{1, 2}) != true {
		t.Fatalf("Encoding not correct: %v", value)
	}
}

func TestParser_ParseLongs__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x00, 0x00, 0x01}

	value, err := p.ParseLongs(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint32{1}) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	encoded = []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}

	value, err = p.ParseLongs(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint32{1}) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseLongs__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}

	value, err := p.ParseLongs(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint32{1, 2}) != true {
		t.Fatalf("Encoding not correct: %v", value)
	}
}

func TestParser_ParseFloats__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x40, 0x49, 0x0f, 0xdb}

	value, err := p.ParseFloats(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expectedResult := []float32{3.14159265}

	for i, v := range value {
		if v < expectedResult[i] ||
			v >= math.Nextafter32(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("Encoding not correct (1): %v", value)
		}
	}
}

func TestParser_ParseFloats__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x40, 0x49, 0x0f, 0xdb, 0x40, 0x2d, 0xf8, 0x54}

	value, err := p.ParseFloats(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	expectedResult := []float32{3.14159265, 2.71828182}

	for i, v := range value {
		if v < expectedResult[i] ||
			v >= math.Nextafter32(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("Encoding not correct (1): %v", value)
		}
	}
}

func TestParser_ParseDoubles__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x40, 0x09, 0x21, 0xfb, 0x53, 0xc8, 0xd4, 0xf1}

	value, err := p.ParseDoubles(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expectedResult := []float64{3.14159265}
	for i, v := range value {
		if v < expectedResult[i] ||
			v >= math.Nextafter(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("Encoding not correct (1): %v", value)
		}
	}
}

func TestParser_ParseDoubles__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x40, 0x09, 0x21, 0xfb, 0x53, 0xc8, 0xd4, 0xf1,
		0x40, 0x05, 0xbf, 0x0a, 0x89, 0xf1, 0xb0, 0xdd}

	value, err := p.ParseDoubles(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	expectedResult := []float64{3.14159265, 2.71828182}

	for i, v := range value {
		if v < expectedResult[i] ||
			v >= math.Nextafter(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("Encoding not correct: %v", value)
		}
	}
}

func TestParser_ParseRationals__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err := p.ParseRationals(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expected := []Rational{
		{Numerator: 1, Denominator: 2},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	encoded = []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err = p.ParseRationals(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expected = []Rational{
		{Numerator: 1, Denominator: 2},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseRationals__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err := p.ParseRationals(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	expected := []Rational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseSignedLongs__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x00, 0x00, 0x01}

	value, err := p.ParseSignedLongs(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []int32{1}) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	encoded = []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}

	value, err = p.ParseSignedLongs(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []int32{1}) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseSignedLongs__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}

	value, err := p.ParseSignedLongs(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	if reflect.DeepEqual(value, []int32{1, 2}) != true {
		t.Fatalf("Encoding not correct: %v", value)
	}
}

func TestParser_ParseSignedRationals__Single(t *testing.T) {
	p := new(Parser)

	encoded := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err := p.ParseSignedRationals(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expected := []SignedRational{
		{Numerator: 1, Denominator: 2},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (1): %v", value)
	}

	encoded = []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err = p.ParseSignedRationals(encoded, 1, TestDefaultByteOrder)
	log.PanicIf(err)

	expected = []SignedRational{
		{Numerator: 1, Denominator: 2},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}

func TestParser_ParseSignedRationals__Multiple(t *testing.T) {
	p := new(Parser)

	encoded := []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x04,
	}

	value, err := p.ParseSignedRationals(encoded, 2, TestDefaultByteOrder)
	log.PanicIf(err)

	expected := []SignedRational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Encoding not correct (2): %v", value)
	}
}
