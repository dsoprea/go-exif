package exif

import (
    "errors"
)

const (
    TypeByte = uint16(1)
    TypeAscii = uint16(2)
    TypeShort = uint16(3)
    TypeLong = uint16(4)
    TypeRational = uint16(5)
    TypeUndefined = uint16(7)
    TypeSignedLong = uint16(9)
    TypeSignedRational = uint16(10)

    // Custom, for our purposes.
    TypeAsciiNoNul = uint16(0xf0)
)

var (
    TypeNames = map[uint16]string {
        TypeByte: "BYTE",
        TypeAscii: "ASCII",
        TypeShort: "SHORT",
        TypeLong: "LONG",
        TypeRational: "RATIONAL",
        TypeUndefined: "UNDEFINED",
        TypeSignedLong: "SLONG",
        TypeSignedRational: "SRATIONAL",

        TypeAsciiNoNul: "_ASCII_NO_NUL",
    }
)
var (
    // ErrCantDetermineTagValueSize is used when we're trying to determine a
    //size for a non-standard/undefined type.
    ErrCantDetermineTagValueSize = errors.New("can not determine tag-value size")

    // ErrNotEnoughData is used when there isn't enough data to accomodate what
    // we're trying to parse (sizeof(type) * unit_count).
    ErrNotEnoughData = errors.New("not enough data for type")

    // ErrWrongType is used when we try to parse anything other than the
    // current type.
    ErrWrongType = errors.New("wrong type, can not parse")

    // ErrUnhandledUnknownTag is used when we try to parse a tag that's
    // recorded as an "unknown" type but not a documented tag (therefore
    // leaving us not knowning how to read it).
    ErrUnhandledUnknownTypedTag = errors.New("not a standard unknown-typed tag")
)


type Rational struct {
    Numerator uint32
    Denominator uint32
}

type SignedRational struct {
    Numerator int32
    Denominator int32
}
