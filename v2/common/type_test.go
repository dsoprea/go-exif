package exifcommon

import (
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestTypeByte_String(t *testing.T) {
	if TypeByte.String() != "BYTE" {
		t.Fatalf("Type name not correct (byte): [%s]", TypeByte.String())
	}
}

func TestTypeAscii_String(t *testing.T) {
	if TypeAscii.String() != "ASCII" {
		t.Fatalf("Type name not correct (ASCII): [%s]", TypeAscii.String())
	}
}

func TestTypeAsciiNoNul_String(t *testing.T) {
	if TypeAsciiNoNul.String() != "_ASCII_NO_NUL" {
		t.Fatalf("Type name not correct (ASCII no-NUL): [%s]", TypeAsciiNoNul.String())
	}
}

func TestTypeShort_String(t *testing.T) {
	if TypeShort.String() != "SHORT" {
		t.Fatalf("Type name not correct (short): [%s]", TypeShort.String())
	}
}

func TestTypeLong_String(t *testing.T) {
	if TypeLong.String() != "LONG" {
		t.Fatalf("Type name not correct (long): [%s]", TypeLong.String())
	}
}

func TestTypeRational_String(t *testing.T) {
	if TypeRational.String() != "RATIONAL" {
		t.Fatalf("Type name not correct (rational): [%s]", TypeRational.String())
	}
}

func TestTypeSignedLong_String(t *testing.T) {
	if TypeSignedLong.String() != "SLONG" {
		t.Fatalf("Type name not correct (signed long): [%s]", TypeSignedLong.String())
	}
}

func TestTypeSignedRational_String(t *testing.T) {
	if TypeSignedRational.String() != "SRATIONAL" {
		t.Fatalf("Type name not correct (signed rational): [%s]", TypeSignedRational.String())
	}
}

func TestTypeByte_Size(t *testing.T) {
	if TypeByte.Size() != 1 {
		t.Fatalf("Type size not correct (byte): (%d)", TypeByte.Size())
	}
}

func TestTypeAscii_Size(t *testing.T) {
	if TypeAscii.Size() != 1 {
		t.Fatalf("Type size not correct (ASCII): (%d)", TypeAscii.Size())
	}
}

func TestTypeAsciiNoNul_Size(t *testing.T) {
	if TypeAsciiNoNul.Size() != 1 {
		t.Fatalf("Type size not correct (ASCII no-NUL): (%d)", TypeAsciiNoNul.Size())
	}
}

func TestTypeShort_Size(t *testing.T) {
	if TypeShort.Size() != 2 {
		t.Fatalf("Type size not correct (short): (%d)", TypeShort.Size())
	}
}

func TestTypeLong_Size(t *testing.T) {
	if TypeLong.Size() != 4 {
		t.Fatalf("Type size not correct (long): (%d)", TypeLong.Size())
	}
}

func TestTypeRational_Size(t *testing.T) {
	if TypeRational.Size() != 8 {
		t.Fatalf("Type size not correct (rational): (%d)", TypeRational.Size())
	}
}

func TestTypeSignedLong_Size(t *testing.T) {
	if TypeSignedLong.Size() != 4 {
		t.Fatalf("Type size not correct (signed long): (%d)", TypeSignedLong.Size())
	}
}

func TestTypeSignedRational_Size(t *testing.T) {
	if TypeSignedRational.Size() != 8 {
		t.Fatalf("Type size not correct (signed rational): (%d)", TypeSignedRational.Size())
	}
}

func TestFormat__Byte(t *testing.T) {
	r := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	s, err := FormatFromBytes(r, TypeByte, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "01 02 03 04 05 06 07 08" {
		t.Fatalf("Format output not correct (bytes): [%s]", s)
	}
}

func TestFormat__Ascii(t *testing.T) {
	r := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	s, err := FormatFromBytes(r, TypeAscii, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "abcdefg" {
		t.Fatalf("Format output not correct (ASCII): [%s]", s)
	}
}

func TestFormat__AsciiNoNul(t *testing.T) {
	r := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	s, err := FormatFromBytes(r, TypeAsciiNoNul, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "abcdefgh" {
		t.Fatalf("Format output not correct (ASCII no-NUL): [%s]", s)
	}
}

func TestFormat__Short(t *testing.T) {
	r := []byte{0, 1, 0, 2}

	s, err := FormatFromBytes(r, TypeShort, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "[1 2]" {
		t.Fatalf("Format output not correct (shorts): [%s]", s)
	}
}

func TestFormat__Long(t *testing.T) {
	r := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	s, err := FormatFromBytes(r, TypeLong, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "[1 2]" {
		t.Fatalf("Format output not correct (longs): [%s]", s)
	}
}

func TestFormat__Rational(t *testing.T) {
	r := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	s, err := FormatFromBytes(r, TypeRational, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "[1/2 3/4]" {
		t.Fatalf("Format output not correct (rationals): [%s]", s)
	}
}

func TestFormat__SignedLong(t *testing.T) {
	r := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	s, err := FormatFromBytes(r, TypeSignedLong, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "[1 2]" {
		t.Fatalf("Format output not correct (signed longs): [%s]", s)
	}
}

func TestFormat__SignedRational(t *testing.T) {
	r := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	s, err := FormatFromBytes(r, TypeSignedRational, false, TestDefaultByteOrder)
	log.PanicIf(err)

	if s != "[1/2 3/4]" {
		t.Fatalf("Format output not correct (signed rationals): [%s]", s)
	}
}

func TestFormat__Undefined(t *testing.T) {
	r := []byte{'a', 'b'}

	_, err := FormatFromBytes(r, TypeUndefined, false, TestDefaultByteOrder)
	if err == nil {
		t.Fatalf("Expected error.")
	} else if err.Error() != "can not determine tag-value size for type (7): [UNDEFINED]" {
		log.Panic(err)
	}
}

func TestTranslateStringToType__TypeUndefined(t *testing.T) {
	_, err := TranslateStringToType(TypeUndefined, "")
	if err == nil {
		t.Fatalf("Expected error.")
	} else if err.Error() != "undefined-type values are not supported" {
		log.Panic(err)
	}
}

func TestTranslateStringToType__TypeByte(t *testing.T) {
	v, err := TranslateStringToType(TypeByte, "02")
	log.PanicIf(err)

	if v != byte(2) {
		t.Fatalf("Translation of string to type not correct (bytes): %v", v)
	}
}

func TestTranslateStringToType__TypeAscii(t *testing.T) {
	v, err := TranslateStringToType(TypeAscii, "abcdefgh")
	log.PanicIf(err)

	if v != "abcdefgh" {
		t.Fatalf("Translation of string to type not correct (ascii): %v", v)
	}
}

func TestTranslateStringToType__TypeAsciiNoNul(t *testing.T) {
	v, err := TranslateStringToType(TypeAsciiNoNul, "abcdefgh")
	log.PanicIf(err)

	if v != "abcdefgh" {
		t.Fatalf("Translation of string to type not correct (ascii no-NUL): %v", v)
	}
}

func TestTranslateStringToType__TypeShort(t *testing.T) {
	v, err := TranslateStringToType(TypeShort, "11")
	log.PanicIf(err)

	if v != uint16(11) {
		t.Fatalf("Translation of string to type not correct (short): %v", v)
	}
}

func TestTranslateStringToType__TypeLong(t *testing.T) {
	v, err := TranslateStringToType(TypeLong, "11")
	log.PanicIf(err)

	if v != uint32(11) {
		t.Fatalf("Translation of string to type not correct (long): %v", v)
	}
}

func TestTranslateStringToType__TypeRational(t *testing.T) {
	v, err := TranslateStringToType(TypeRational, "11/22")
	log.PanicIf(err)

	r := v.(Rational)

	if r.Numerator != 11 || r.Denominator != 22 {
		t.Fatalf("Translation of string to type not correct (rational): %v", r)
	}
}

func TestTranslateStringToType__TypeSignedLong(t *testing.T) {
	v, err := TranslateStringToType(TypeSignedLong, "11")
	log.PanicIf(err)

	if v != int32(11) {
		t.Fatalf("Translation of string to type not correct (signed long): %v", v)
	}
}

func TestTranslateStringToType__TypeSignedRational(t *testing.T) {
	v, err := TranslateStringToType(TypeSignedRational, "11/22")
	log.PanicIf(err)

	r := v.(SignedRational)

	if r.Numerator != 11 || r.Denominator != 22 {
		t.Fatalf("Translation of string to type not correct (signed rational): %v", r)
	}
}

func TestTranslateStringToType__InvalidType(t *testing.T) {
	_, err := TranslateStringToType(99, "11/22")
	if err == nil {
		t.Fatalf("Expected error for invalid type.")
	} else if err.Error() != "from-string encoding for type not supported; this shouldn't happen: []" {
		log.Panic(err)
	}
}

//     } else if tagType == TypeLong {
//         n, err := strconv.ParseUint(valueString, 10, 32)
//         log.PanicIf(err)

//         return uint32(n), nil
//     } else if tagType == TypeRational {
//         parts := strings.SplitN(valueString, "/", 2)

//         numerator, err := strconv.ParseUint(parts[0], 10, 32)
//         log.PanicIf(err)

//         denominator, err := strconv.ParseUint(parts[1], 10, 32)
//         log.PanicIf(err)

//         return Rational{
//             Numerator:   uint32(numerator),
//             Denominator: uint32(denominator),
//         }, nil
//     } else if tagType == TypeSignedLong {
//         n, err := strconv.ParseInt(valueString, 10, 32)
//         log.PanicIf(err)

//         return int32(n), nil
//     } else if tagType == TypeSignedRational {
//         parts := strings.SplitN(valueString, "/", 2)

//         numerator, err := strconv.ParseInt(parts[0], 10, 32)
//         log.PanicIf(err)

//         denominator, err := strconv.ParseInt(parts[1], 10, 32)
//         log.PanicIf(err)

//         return SignedRational{
//             Numerator:   int32(numerator),
//             Denominator: int32(denominator),
//         }, nil
//     }

//     log.Panicf("from-string encoding for type not supported; this shouldn't happen: [%s]", tagType.String())
//     return nil, nil
// }
