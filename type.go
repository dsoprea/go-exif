package exif

import (
    "bytes"
    "errors"
    "fmt"
    "strconv"
    "strings"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

type TagTypePrimitive uint16

func (tagType TagTypePrimitive) Size() int {
    if tagType == TypeByte {
        return 1
    } else if tagType == TypeAscii || tagType == TypeAsciiNoNul {
        return 1
    } else if tagType == TypeShort {
        return 2
    } else if tagType == TypeLong {
        return 4
    } else if tagType == TypeRational {
        return 8
    } else if tagType == TypeSignedLong {
        return 4
    } else if tagType == TypeSignedRational {
        return 8
    } else {
        log.Panicf("can not determine tag-value size for type (%d): [%s]", tagType, TypeNames[tagType])

        // Never called.
        return 0
    }
}

const (
    TypeByte           TagTypePrimitive = 1
    TypeAscii                           = 2
    TypeShort                           = 3
    TypeLong                            = 4
    TypeRational                        = 5
    TypeUndefined                       = 7
    TypeSignedLong                      = 9
    TypeSignedRational                  = 10

    // TypeAsciiNoNul is just a pseudo-type, for our own purposes.
    TypeAsciiNoNul = 0xf0
)

var (
    typeLogger = log.NewLogger("exif.type")
)

var (
    // TODO(dustin): Rename TypeNames() to typeNames() and add getter.
    TypeNames = map[TagTypePrimitive]string{
        TypeByte:           "BYTE",
        TypeAscii:          "ASCII",
        TypeShort:          "SHORT",
        TypeLong:           "LONG",
        TypeRational:       "RATIONAL",
        TypeUndefined:      "UNDEFINED",
        TypeSignedLong:     "SLONG",
        TypeSignedRational: "SRATIONAL",

        TypeAsciiNoNul: "_ASCII_NO_NUL",
    }

    TypeNamesR = map[string]TagTypePrimitive{}
)

var (
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
    Numerator   uint32
    Denominator uint32
}

type SignedRational struct {
    Numerator   int32
    Denominator int32
}

type TagType struct {
    tagType   TagTypePrimitive
    name      string
    byteOrder binary.ByteOrder
}

func NewTagType(tagType TagTypePrimitive, byteOrder binary.ByteOrder) TagType {
    name, found := TypeNames[tagType]
    if found == false {
        log.Panicf("tag-type not valid: 0x%04x", tagType)
    }

    return TagType{
        tagType:   tagType,
        name:      name,
        byteOrder: byteOrder,
    }
}

func (tt TagType) String() string {
    return fmt.Sprintf("TagType<NAME=[%s]>", tt.name)
}

func (tt TagType) Name() string {
    return tt.name
}

func (tt TagType) Type() TagTypePrimitive {
    return tt.tagType
}

func (tt TagType) ByteOrder() binary.ByteOrder {
    return tt.byteOrder
}

// DEPRECATED(dustin): `(TagTypePrimitive).Size()` should be used, directly.
func (tt TagType) Size() int {
    return tt.Type().Size()
}

// DEPRECATED(dustin): `(TagTypePrimitive).Size()` should be used, directly.
func TagTypeSize(tagType TagTypePrimitive) int {
    return tagType.Size()
}

// valueIsEmbedded will return a boolean indicating whether the value should be
// found directly within the IFD entry or an offset to somewhere else.
func (tt TagType) valueIsEmbedded(unitCount uint32) bool {
    return (tt.tagType.Size() * int(unitCount)) <= 4
}

func (tt TagType) readRawEncoded(valueContext ValueContext) (rawBytes []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    unitSizeRaw := uint32(tt.tagType.Size())

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        byteLength := unitSizeRaw * valueContext.UnitCount()
        return valueContext.RawValueOffset()[:byteLength], nil
    } else {
        return valueContext.AddressableData()[valueContext.ValueOffset() : valueContext.ValueOffset()+valueContext.UnitCount()*unitSizeRaw], nil
    }
}

func (tt TagType) ParseBytes(data []byte, unitCount uint32) (value []uint8, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeByte {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = []uint8(data[:count])

    return value, nil
}

// ParseAscii returns a string and auto-strips the trailing NUL character.
func (tt TagType) ParseAscii(data []byte, unitCount uint32) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeAscii && tt.tagType != TypeAsciiNoNul {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    if len(data) == 0 || data[count-1] != 0 {
        s := string(data[:count])
        typeLogger.Warningf(nil, "ascii not terminated with nul as expected: [%v]", s)

        return s, nil
    } else {
        // Auto-strip the NUL from the end. It serves no purpose outside of
        // encoding semantics.

        return string(data[:count-1]), nil
    }
}

// ParseAsciiNoNul returns a string without any consideration for a trailing NUL
// character.
func (tt TagType) ParseAsciiNoNul(data []byte, unitCount uint32) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeAscii && tt.tagType != TypeAsciiNoNul {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    return string(data[:count]), nil
}

func (tt TagType) ParseShorts(data []byte, unitCount uint32) (value []uint16, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeShort {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]uint16, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
            value[i] = binary.BigEndian.Uint16(data[i*2:])
        } else {
            value[i] = binary.LittleEndian.Uint16(data[i*2:])
        }
    }

    return value, nil
}

func (tt TagType) ParseLongs(data []byte, unitCount uint32) (value []uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeLong {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]uint32, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
            value[i] = binary.BigEndian.Uint32(data[i*4:])
        } else {
            value[i] = binary.LittleEndian.Uint32(data[i*4:])
        }
    }

    return value, nil
}

func (tt TagType) ParseRationals(data []byte, unitCount uint32) (value []Rational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeRational {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]Rational, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
            value[i].Numerator = binary.BigEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.BigEndian.Uint32(data[i*8+4:])
        } else {
            value[i].Numerator = binary.LittleEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.LittleEndian.Uint32(data[i*8+4:])
        }
    }

    return value, nil
}

func (tt TagType) ParseSignedLongs(data []byte, unitCount uint32) (value []int32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeSignedLong {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    b := bytes.NewBuffer(data)

    value = make([]int32, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
            err := binary.Read(b, binary.BigEndian, &value[i])
            log.PanicIf(err)
        } else {
            err := binary.Read(b, binary.LittleEndian, &value[i])
            log.PanicIf(err)
        }
    }

    return value, nil
}

func (tt TagType) ParseSignedRationals(data []byte, unitCount uint32) (value []SignedRational, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType != TypeSignedRational {
        log.Panic(ErrWrongType)
    }

    count := int(unitCount)

    if len(data) < (tt.tagType.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    b := bytes.NewBuffer(data)

    value = make([]SignedRational, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading BYTE value (embedded).")

        // In this case, the bytes normally used for the offset are actually
        // data.

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseBytes(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading BYTE value (at offset).")

        value, err = tt.ParseBytes(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading ASCII value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseAscii(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading ASCII value (at offset).")

        value, err = tt.ParseAscii(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
        log.PanicIf(err)
    }

    return value, nil
}

func (tt TagType) ReadAsciiNoNulValue(valueContext ValueContext) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading ASCII value (no-nul; embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseAsciiNoNul(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading ASCII value (no-nul; at offset).")

        value, err = tt.ParseAsciiNoNul(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading SHORT value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseShorts(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SHORT value (at offset).")

        value, err = tt.ParseShorts(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading LONG value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseLongs(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading LONG value (at offset).")

        value, err = tt.ParseLongs(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading RATIONAL value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseRationals(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading RATIONAL value (at offset).")

        value, err = tt.ParseRationals(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading SLONG value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseSignedLongs(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SLONG value (at offset).")

        value, err = tt.ParseSignedLongs(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
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

    if tt.valueIsEmbedded(valueContext.UnitCount()) == true {
        typeLogger.Debugf(nil, "Reading SRATIONAL value (embedded).")

        byteLength := uint32(tt.tagType.Size()) * valueContext.UnitCount()
        rawValue := valueContext.RawValueOffset()[:byteLength]

        value, err = tt.ParseSignedRationals(rawValue, valueContext.UnitCount())
        log.PanicIf(err)
    } else {
        typeLogger.Debugf(nil, "Reading SRATIONAL value (at offset).")

        value, err = tt.ParseSignedRationals(valueContext.AddressableData()[valueContext.ValueOffset():], valueContext.UnitCount())
        log.PanicIf(err)
    }

    return value, nil
}

// ResolveAsString resolves the given value and returns a flat string.
//
// Where the type is not ASCII, `justFirst` indicates whether to just stringify
// the first item in the slice (or return an empty string if the slice is
// empty).
//
// Since this method lacks the information to process unknown-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `UndefinedValue()`.
func (tt TagType) ResolveAsString(valueContext ValueContext, justFirst bool) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    rawBytes, err := tt.readRawEncoded(valueContext)
    log.PanicIf(err)

    valueString, err := tt.Format(rawBytes, justFirst)
    log.PanicIf(err)

    return valueString, nil
}

// Format returns a stringified value for the given bytes. Automatically
// calculates count based on type size.
func (tt TagType) Format(rawBytes []byte, justFirst bool) (value string, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    // TODO(dustin): !! Add tests

    typeId := tt.Type()
    typeSize := typeId.Size()

    if len(rawBytes)%typeSize != 0 {
        log.Panicf("byte-count (%d) does not align for [%s] type with a size of (%d) bytes", len(rawBytes), TypeNames[typeId], typeSize)
    }

    // unitCount is the calculated unit-count. This should equal the original
    // value from the tag (pre-resolution).
    unitCount := uint32(len(rawBytes) / typeSize)

    // Truncate the items if it's not bytes or a string and we just want the first.

    valueSuffix := ""
    if justFirst == true && unitCount > 1 && typeId != TypeByte && typeId != TypeAscii && typeId != TypeAsciiNoNul {
        unitCount = 1
        valueSuffix = "..."
    }

    if typeId == TypeByte {
        items, err := tt.ParseBytes(rawBytes, unitCount)
        log.PanicIf(err)

        return DumpBytesToString(items), nil
    } else if typeId == TypeAscii {
        phrase, err := tt.ParseAscii(rawBytes, unitCount)
        log.PanicIf(err)

        return phrase, nil
    } else if typeId == TypeAsciiNoNul {
        phrase, err := tt.ParseAsciiNoNul(rawBytes, unitCount)
        log.PanicIf(err)

        return phrase, nil
    } else if typeId == TypeShort {
        items, err := tt.ParseShorts(rawBytes, unitCount)
        log.PanicIf(err)

        if len(items) > 0 {
            if justFirst == true {
                return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
            } else {
                return fmt.Sprintf("%v", items), nil
            }
        } else {
            return "", nil
        }
    } else if typeId == TypeLong {
        items, err := tt.ParseLongs(rawBytes, unitCount)
        log.PanicIf(err)

        if len(items) > 0 {
            if justFirst == true {
                return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
            } else {
                return fmt.Sprintf("%v", items), nil
            }
        } else {
            return "", nil
        }
    } else if typeId == TypeRational {
        items, err := tt.ParseRationals(rawBytes, unitCount)
        log.PanicIf(err)

        if len(items) > 0 {
            parts := make([]string, len(items))
            for i, r := range items {
                parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
            }

            if justFirst == true {
                return fmt.Sprintf("%v%s", parts[0], valueSuffix), nil
            } else {
                return fmt.Sprintf("%v", parts), nil
            }
        } else {
            return "", nil
        }
    } else if typeId == TypeSignedLong {
        items, err := tt.ParseSignedLongs(rawBytes, unitCount)
        log.PanicIf(err)

        if len(items) > 0 {
            if justFirst == true {
                return fmt.Sprintf("%v%s", items[0], valueSuffix), nil
            } else {
                return fmt.Sprintf("%v", items), nil
            }
        } else {
            return "", nil
        }
    } else if typeId == TypeSignedRational {
        items, err := tt.ParseSignedRationals(rawBytes, unitCount)
        log.PanicIf(err)

        parts := make([]string, len(items))
        for i, r := range items {
            parts[i] = fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
        }

        if len(items) > 0 {
            if justFirst == true {
                return fmt.Sprintf("%v%s", parts[0], valueSuffix), nil
            } else {
                return fmt.Sprintf("%v", parts), nil
            }
        } else {
            return "", nil
        }
    } else {
        // Affects only "unknown" values, in general.
        log.Panicf("value of type (%d) [%s] can not be formatted into string", typeId, tt)

        // Never called.
        return "", nil
    }
}

// Value knows how to resolve the given value.
//
// Since this method lacks the information to process unknown-type tags (e.g.
// byte-order, tag-ID, IFD type), it will return an error if attempted. See
// `UndefinedValue()`.
func (tt TagType) Resolve(valueContext ValueContext) (value interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    typeId := tt.Type()

    if typeId == TypeByte {
        value, err = tt.ReadByteValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeAscii {
        value, err = tt.ReadAsciiValue(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeAsciiNoNul {
        value, err = tt.ReadAsciiNoNulValue(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeShort {
        value, err = tt.ReadShortValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeLong {
        value, err = tt.ReadLongValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeRational {
        value, err = tt.ReadRationalValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeSignedLong {
        value, err = tt.ReadSignedLongValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeSignedRational {
        value, err = tt.ReadSignedRationalValues(valueContext)
        log.PanicIf(err)
    } else if typeId == TypeUndefined {
        log.Panicf("will not parse unknown-type value: %v", tt)

        // Never called.
        return nil, nil
    } else {
        log.Panicf("value of type (%d) [%s] is unparseable", typeId, tt)

        // Never called.
        return nil, nil
    }

    return value, nil
}

// Encode knows how to encode the given value to a byte slice.
func (tt TagType) Encode(value interface{}) (encoded []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ve := NewValueEncoder(tt.byteOrder)

    ed, err := ve.EncodeWithType(tt, value)
    log.PanicIf(err)

    return ed.Encoded, err
}

func (tt TagType) FromString(valueString string) (value interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if tt.tagType == TypeUndefined {
        // TODO(dustin): Circle back to this.
        log.Panicf("undefined-type values are not supported")
    }

    if tt.tagType == TypeByte {
        return []byte(valueString), nil
    } else if tt.tagType == TypeAscii || tt.tagType == TypeAsciiNoNul {
        // Whether or not we're putting an NUL on the end is only relevant for
        // byte-level encoding. This function really just supports a user
        // interface.

        return valueString, nil
    } else if tt.tagType == TypeShort {
        n, err := strconv.ParseUint(valueString, 10, 16)
        log.PanicIf(err)

        return uint16(n), nil
    } else if tt.tagType == TypeLong {
        n, err := strconv.ParseUint(valueString, 10, 32)
        log.PanicIf(err)

        return uint32(n), nil
    } else if tt.tagType == TypeRational {
        parts := strings.SplitN(valueString, "/", 2)

        numerator, err := strconv.ParseUint(parts[0], 10, 32)
        log.PanicIf(err)

        denominator, err := strconv.ParseUint(parts[1], 10, 32)
        log.PanicIf(err)

        return Rational{
            Numerator:   uint32(numerator),
            Denominator: uint32(denominator),
        }, nil
    } else if tt.tagType == TypeSignedLong {
        n, err := strconv.ParseInt(valueString, 10, 32)
        log.PanicIf(err)

        return int32(n), nil
    } else if tt.tagType == TypeSignedRational {
        parts := strings.SplitN(valueString, "/", 2)

        numerator, err := strconv.ParseInt(parts[0], 10, 32)
        log.PanicIf(err)

        denominator, err := strconv.ParseInt(parts[1], 10, 32)
        log.PanicIf(err)

        return SignedRational{
            Numerator:   int32(numerator),
            Denominator: int32(denominator),
        }, nil
    }

    log.Panicf("from-string encoding for type not supported; this shouldn't happen: (%d)", tt.Type())
    return nil, nil
}

func init() {
    for typeId, typeName := range TypeNames {
        TypeNamesR[typeName] = typeId
    }
}
