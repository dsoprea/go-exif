package exif

import (
    "bytes"
    "errors"
    "fmt"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
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
    typeLogger = log.NewLogger("exif.type")
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


const (
    BigEndianByteOrder = iota
    LittleEndianByteOrder = iota
)

type IfdByteOrder int

func (ibo IfdByteOrder) IsBigEndian() bool {
    return ibo == BigEndianByteOrder
}

func (ibo IfdByteOrder) IsLittleEndian() bool {
    return ibo == LittleEndianByteOrder
}


type Rational struct {
    Numerator uint32
    Denominator uint32
}

type SignedRational struct {
    Numerator int32
    Denominator int32
}


type TagType struct {
    tagType uint16
    name string
    byteOrder IfdByteOrder
}

func NewTagType(tagType uint16, byteOrder IfdByteOrder) TagType {
    name, found := TypeNames[tagType]
    if found == false {
        log.Panicf("tag-type not valid: 0x%04x", tagType)
    }

    return TagType{
        tagType: tagType,
        name: name,
        byteOrder: byteOrder,
    }
}

func (tt TagType) String() string {
    return fmt.Sprintf("TagType<NAME=[%s]>", tt.name)
}

func (tt TagType) Name() string {
    return tt.name
}

func (tt TagType) Type() uint16 {
    return tt.tagType
}

func (tt TagType) ByteOrder() IfdByteOrder {
    return tt.byteOrder
}


func (tt TagType) Size() int {
    if tt.tagType == TypeByte {
        return 1
    } else if tt.tagType == TypeAscii || tt.tagType == TypeAsciiNoNul {
        return 1
    } else if tt.tagType == TypeShort {
        return 2
    } else if tt.tagType == TypeLong {
        return 4
    } else if tt.tagType == TypeRational {
        return 8
    } else if tt.tagType == TypeSignedLong {
        return 4
    } else if tt.tagType == TypeSignedRational {
        return 8
    } else {
        log.Panic(ErrCantDetermineTagValueSize)

        // Never called.
        return 0
    }
}

// ValueIsEmbedded will return a boolean indicating whether the value should be
// found directly within the IFD entry or an offset to somewhere else.
func (tt TagType) ValueIsEmbedded(unitCount uint32) bool {
    return (tt.Size() * int(unitCount)) <= 4
}

func (tt TagType) ParseBytes(data []byte, rawCount uint32) (value []uint8, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeByte {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]uint8, count)
    for i := 0; i < count; i++ {
        value[i] = uint8(data[i])
    }

    return value, nil
}

func (tt TagType) ParseAscii(data []byte, rawCount uint32) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeAscii && tt.tagType != TypeAsciiNoNul {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    return string(data[:count]), nil
}

func (tt TagType) ParseShorts(data []byte, rawCount uint32) (value []uint16, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeShort {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]uint16, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder.IsBigEndian() {
            value[i] = binary.BigEndian.Uint16(data[i*2:])
        } else {
            value[i] = binary.LittleEndian.Uint16(data[i*2:])
        }
    }

    return value, nil
}

func (tt TagType) ParseLongs(data []byte, rawCount uint32) (value []uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeLong {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]uint32, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder.IsBigEndian() {
            value[i] = binary.BigEndian.Uint32(data[i*4:])
        } else {
            value[i] = binary.LittleEndian.Uint32(data[i*4:])
        }
    }

    return value, nil
}

func (tt TagType) ParseRationals(data []byte, rawCount uint32) (value []Rational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeRational {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]Rational, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder.IsBigEndian() {
            value[i].Numerator = binary.BigEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.BigEndian.Uint32(data[i*8 + 4:])
        } else {
            value[i].Numerator = binary.LittleEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.LittleEndian.Uint32(data[i*8 + 4:])
        }
    }

    return value, nil
}

func (tt TagType) ParseSignedLongs(data []byte, rawCount uint32) (value []int32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeSignedLong {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    b := bytes.NewBuffer(data)

    value = make([]int32, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder.IsBigEndian() {
            err := binary.Read(b, binary.BigEndian, &value[i])
            log.PanicIf(err)
        } else {
            err := binary.Read(b, binary.LittleEndian, &value[i])
            log.PanicIf(err)
        }
    }

    return value, nil
}

func (tt TagType) ParseSignedRationals(data []byte, rawCount uint32) (value []SignedRational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeSignedRational {
        log.Panic(ErrWrongType)
    }

    count := int(rawCount)

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    b := bytes.NewBuffer(data)

    value = make([]SignedRational, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder.IsBigEndian() {
            err = binary.Read(b, binary.BigEndian, &value[i].Numerator)
            log.PanicIf(err)

            err = binary.Read(b, binary.BigEndian, &value[i].Denominator)
            log.PanicIf(err)
        } else {
            err = binary.Read(b, binary.LittleEndian, &value[i].Numerator)
            log.PanicIf(err)

            err = binary.Read(b, binary.LittleEndian, &value[i].Denominator)
            log.PanicIf(err)
        }
    }

    return value, nil
}

func (tt TagType) ReadByteValues(valueContext ValueContext) (value []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading BYTE value (embedded).")

        value, err = tt.ParseBytes(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading BYTE value (at offset).")

        value, err = tt.ParseBytes(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadAsciiValue(valueContext ValueContext) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    value, err = tt.ReadAsciiNoNulValue(valueContext)
    log.PanicIf(err)

    len_ := len(value)
    if len_ == 0 || value[len_ - 1] != 0 {
        typeLogger.Warningf(nil, "ascii value not terminated with nul: [%s]", value)
        return value, nil
    } else {
        return value[:len_ - 1], nil
    }
}

func (tt TagType) ReadAsciiNoNulValue(valueContext ValueContext) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading ASCII value (no-nul; embedded).")

        value, err = tt.ParseAscii(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading ASCII value (no-nul; at offset).")

        value, err = tt.ParseAscii(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadShortValues(valueContext ValueContext) (value []uint16, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading SHORT value (embedded).")

        value, err = tt.ParseShorts(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SHORT value (at offset).")

        value, err = tt.ParseShorts(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadLongValues(valueContext ValueContext) (value []uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading LONG value (embedded).")

        value, err = tt.ParseLongs(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading LONG value (at offset).")

        value, err = tt.ParseLongs(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadRationalValues(valueContext ValueContext) (value []Rational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading RATIONAL value (embedded).")

        value, err = tt.ParseRationals(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading RATIONAL value (at offset).")

        value, err = tt.ParseRationals(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadSignedLongValues(valueContext ValueContext) (value []int32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading SLONG value (embedded).")

        value, err = tt.ParseSignedLongs(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SLONG value (at offset).")

        value, err = tt.ParseSignedLongs(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadSignedRationalValues(valueContext ValueContext) (value []SignedRational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeLogger.Debugf(nil, "Reading SRATIONAL value (embedded).")

        value, err = tt.ParseSignedRationals(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SRATIONAL value (at offset).")

        value, err = tt.ParseSignedRationals(valueContext.RawExif[valueContext.ValueOffset:], valueContext.UnitCount)
        log.PanicIf(err)
    }

    return value, nil
}

// ValueString extracts and parses the given value, and returns a flat string.
// Where the type is not ASCII, `justFirst` indicates whether to just stringify
// the first item in the slice (or return an empty string if the slice is
// empty).
func (tt TagType) ValueString(valueContext ValueContext, justFirst bool) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.Type() == TypeByte {
        raw, err := tt.ReadByteValues(valueContext)
        log.PanicIf(err)

        if justFirst == false {
            return fmt.Sprintf("%v", raw), nil
        } else if valueContext.UnitCount > 0 {
            return fmt.Sprintf("%v", raw[0]), nil
        } else {
            return "", nil
        }
    } else if tt.Type() == TypeAscii {
        raw, err := tt.ReadAsciiValue(valueContext)
        log.PanicIf(err)

        return fmt.Sprintf("%s", raw), nil
    } else if tt.Type() == TypeAsciiNoNul {
        raw, err := tt.ReadAsciiNoNulValue(valueContext)
        log.PanicIf(err)

        return fmt.Sprintf("%s", raw), nil
    } else if tt.Type() == TypeShort {
        raw, err := tt.ReadShortValues(valueContext)
        log.PanicIf(err)

        if justFirst == false {
            return fmt.Sprintf("%v", raw), nil
        } else if valueContext.UnitCount > 0 {
            return fmt.Sprintf("%v", raw[0]), nil
        } else {
            return "", nil
        }
    } else if tt.Type() == TypeLong {
        raw, err := tt.ReadLongValues(valueContext)
        log.PanicIf(err)

        if justFirst == false {
            return fmt.Sprintf("%v", raw), nil
        } else if valueContext.UnitCount > 0 {
            return fmt.Sprintf("%v", raw[0]), nil
        } else {
            return "", nil
        }
    } else if tt.Type() == TypeRational {
        raw, err := tt.ReadRationalValues(valueContext)
        log.PanicIf(err)

        parts := make([]string, len(raw))
        for i, r := range raw {
            parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
        }

        if justFirst == false {
            return fmt.Sprintf("%v", parts), nil
        } else if valueContext.UnitCount > 0 {
            return parts[0], nil
        } else {
            return "", nil
        }
    } else if tt.Type() == TypeSignedLong {
        raw, err := tt.ReadSignedLongValues(valueContext)
        log.PanicIf(err)

        if justFirst == false {
            return fmt.Sprintf("%v", raw), nil
        } else if valueContext.UnitCount > 0 {
            return fmt.Sprintf("%v", raw[0]), nil
        } else {
            return "", nil
        }
    } else if tt.Type() == TypeSignedRational {
        raw, err := tt.ReadSignedRationalValues(valueContext)
        log.PanicIf(err)

        parts := make([]string, len(raw))
        for i, r := range raw {
            parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
        }

        if justFirst == false {
            return fmt.Sprintf("%v", raw), nil
        } else if valueContext.UnitCount > 0 {
            return parts[0], nil
        } else {
            return "", nil
        }
    } else {
        log.Panicf("value of type (%d) [%s] is unparseable", tt.Type(), tt)

        // Never called.
        return "", nil
    }
}

// UndefinedValue returns the value for a tag of "undefined" type.
func UndefinedValue(indexedIfdName string, tagId uint16, valueContext ValueContext, byteOrder IfdByteOrder) (value interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if indexedIfdName == IfdName(IfdExif, 0) {
        if tagId == 0x9000 {
            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            value, err = tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return value, nil
        } else if tagId == 0xa000 {
            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            value, err = tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return value, nil
        }
    } else if indexedIfdName == IfdName(IfdGps, 0) {
        if tagId == 0x001c {
            // GPSAreaInformation

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            value, err = tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return value, nil
        } else if tagId == 0x001b {
            // GPSProcessingMethod

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            value, err = tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return value, nil
        }
    }

// TODO(dustin): !! Still need to do:
//
// complex: 0xa302, 0xa20c, 0x8828
// long: 0xa301, 0xa300
// bytes: 0x927c, 0x9101 (probably, but not certain)
// other: 0x9286 (simple, but needs some processing)

    // 0xa40b is device-specific and unhandled.


    log.Panic(ErrUnhandledUnknownTypedTag)
    return nil, nil
}
