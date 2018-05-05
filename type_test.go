package exif

import (
    "testing"
    "bytes"
    "fmt"
    "reflect"

    "github.com/dsoprea/go-logging"
)

func Test_TagType_EncodeDecode_Byte(t *testing.T) {
    tt := NewTagType(TypeByte, TestDefaultByteOrder)

    data := []byte { 0x11, 0x22, 0x33, 0x44, 0x55 }

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

func Test_TagType_EncodeDecode_Ascii(t *testing.T) {
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

func Test_TagType_EncodeDecode_Shorts(t *testing.T) {
    tt := NewTagType(TypeShort, TestDefaultByteOrder)

    data := []uint16 { 0x11, 0x22, 0x33 }

    encoded, err := tt.Encode(data)
    log.PanicIf(err)

    if bytes.Compare(encoded, []byte { 0x00, 0x11, 0x00, 0x22, 0x00, 0x33 }) != 0 {
        t.Fatalf("Data not encoded correctly.")
    }

    restored, err := tt.ParseShorts(encoded, uint32(len(data)))
    log.PanicIf(err)

    if reflect.DeepEqual(restored, data) != true {
        t.Fatalf("Data not decoded correctly.")
    }
}

func Test_TagType_EncodeDecode_Long(t *testing.T) {
    tt := NewTagType(TypeLong, TestDefaultByteOrder)

    data := []uint32 { 0x11, 0x22, 0x33 }

    encoded, err := tt.Encode(data)
    log.PanicIf(err)

    if bytes.Compare(encoded, []byte { 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33 }) != 0 {
        t.Fatalf("Data not encoded correctly.")
    }

    restored, err := tt.ParseLongs(encoded, uint32(len(data)))
    log.PanicIf(err)

    if reflect.DeepEqual(restored, data) != true {
        t.Fatalf("Data not decoded correctly.")
    }
}

func Test_TagType_EncodeDecode_Rational(t *testing.T) {
    tt := NewTagType(TypeRational, TestDefaultByteOrder)

    data := []Rational {
        Rational{ Numerator: 0x11, Denominator: 0x22 },
        Rational{ Numerator: 0x33, Denominator: 0x44 },
    }

    encoded, err := tt.Encode(data)
    log.PanicIf(err)

    if bytes.Compare(encoded, []byte { 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0x00, 0x44 }) != 0 {
        t.Fatalf("Data not encoded correctly.")
    }

    restored, err := tt.ParseRationals(encoded, uint32(len(data)))
    log.PanicIf(err)

    if reflect.DeepEqual(restored, data) != true {
        t.Fatalf("Data not decoded correctly.")
    }
}

func Test_TagType_EncodeDecode_SignedLong(t *testing.T) {
    tt := NewTagType(TypeSignedLong, TestDefaultByteOrder)

    data := []int32 { 0x11, 0x22, 0x33 }

    encoded, err := tt.Encode(data)
    log.PanicIf(err)

    if bytes.Compare(encoded, []byte { 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33 }) != 0 {
        t.Fatalf("Data not encoded correctly.")
    }

    restored, err := tt.ParseSignedLongs(encoded, uint32(len(data)))
    log.PanicIf(err)

    if reflect.DeepEqual(restored, data) != true {
        t.Fatalf("Data not decoded correctly.")
    }
}

func Test_TagType_EncodeDecode_SignedRational(t *testing.T) {
    tt := NewTagType(TypeSignedRational, TestDefaultByteOrder)

    data := []SignedRational {
        SignedRational{ Numerator: 0x11, Denominator: 0x22 },
        SignedRational{ Numerator: 0x33, Denominator: 0x44 },
    }

    encoded, err := tt.Encode(data)
    log.PanicIf(err)

    if bytes.Compare(encoded, []byte { 0x00, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x22, 0x00, 0x00, 0x00, 0x33, 0x00, 0x00, 0x00, 0x44 }) != 0 {
        t.Fatalf("Data not encoded correctly.")
    }

    restored, err := tt.ParseSignedRationals(encoded, uint32(len(data)))
    log.PanicIf(err)

    if reflect.DeepEqual(restored, data) != true {
        t.Fatalf("Data not decoded correctly.")
    }
}

func Test_TagType_EncodeDecode_AsciiNoNul(t *testing.T) {
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
