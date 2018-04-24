package exif

import (
    "bytes"
    "fmt"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    typeDecodeLogger = log.NewLogger("exif.type_decode")
)


type TagType struct {
    tagType uint16
    name string
    byteOrder binary.ByteOrder
}

func NewTagType(tagType uint16, byteOrder binary.ByteOrder) TagType {
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

func (tt TagType) ByteOrder() binary.ByteOrder {
    return tt.byteOrder
}

func (tt TagType) Size() int {
    return TagTypeSize(tt.Type())
}

func TagTypeSize(tagType int) int {
    if tagType == TypeByte {
        return 1
    } else if tagType == TypeAscii || tt.tagType == TypeAsciiNoNul {
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

// ValueIsEmbedded will return a boolean indicating whether the value should be
// found directly within the IFD entry or an offset to somewhere else.
func (tt TagType) ValueIsEmbedded(unitCount uint32) bool {
    return (tt.Size() * int(unitCount)) <= 4
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

    if len(data) < (tt.Size() * count) {
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

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    if len(data) == 0 || data[count - 1] != 0 {
        typeDecodeLogger.Warningf(nil, "ascii not terminated with nul")
        return string(data[:count]), nil
    } else {
        return string(data[:count - 1]), nil
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

    if len(data) < (tt.Size() * count) {
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

    if len(data) < (tt.Size() * count) {
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

    if len(data) < (tt.Size() * count) {
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

    if len(data) < (tt.Size() * count) {
        log.Panic(ErrNotEnoughData)
    }

    value = make([]Rational, count)
    for i := 0; i < count; i++ {
        if tt.byteOrder == binary.BigEndian {
            value[i].Numerator = binary.BigEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.BigEndian.Uint32(data[i*8 + 4:])
        } else {
            value[i].Numerator = binary.LittleEndian.Uint32(data[i*8:])
            value[i].Denominator = binary.LittleEndian.Uint32(data[i*8 + 4:])
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

    if len(data) < (tt.Size() * count) {
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

    if len(data) < (tt.Size() * count) {
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

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeDecodeLogger.Debugf(nil, "Reading BYTE value (embedded).")

        // In this case, the bytes normally used for the offset are actually
        // data.
        value, err = tt.ParseBytes(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading BYTE value (at offset).")

        value, err = tt.ParseBytes(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeDecodeLogger.Debugf(nil, "Reading ASCII value (no-nul; embedded).")

        value, err = tt.ParseAscii(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading ASCII value (no-nul; at offset).")

        value, err = tt.ParseAscii(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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

    if tt.ValueIsEmbedded(valueContext.UnitCount) == true {
        typeDecodeLogger.Debugf(nil, "Reading ASCII value (no-nul; embedded).")

        value, err = tt.ParseAsciiNoNul(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading ASCII value (no-nul; at offset).")

        value, err = tt.ParseAsciiNoNul(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
        typeDecodeLogger.Debugf(nil, "Reading SHORT value (embedded).")

        value, err = tt.ParseShorts(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading SHORT value (at offset).")

        value, err = tt.ParseShorts(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
        typeDecodeLogger.Debugf(nil, "Reading LONG value (embedded).")

        value, err = tt.ParseLongs(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading LONG value (at offset).")

        value, err = tt.ParseLongs(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
        typeDecodeLogger.Debugf(nil, "Reading RATIONAL value (embedded).")

        value, err = tt.ParseRationals(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading RATIONAL value (at offset).")

        value, err = tt.ParseRationals(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
        typeDecodeLogger.Debugf(nil, "Reading SLONG value (embedded).")

        value, err = tt.ParseSignedLongs(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading SLONG value (at offset).")

        value, err = tt.ParseSignedLongs(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
        typeDecodeLogger.Debugf(nil, "Reading SRATIONAL value (embedded).")

        value, err = tt.ParseSignedRationals(valueContext.RawValueOffset, valueContext.UnitCount)
        log.PanicIf(err)
    } else {
        typeDecodeLogger.Debugf(nil, "Reading SRATIONAL value (at offset).")

        value, err = tt.ParseSignedRationals(valueContext.AddressableData[valueContext.ValueOffset:], valueContext.UnitCount)
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
func UndefinedValue(indexedIfdName string, tagId uint16, valueContext ValueContext, byteOrder binary.ByteOrder) (value interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    typeDecodeLogger.Debugf(nil, "UndefinedValue: IFD=[%s] TAG-ID=(0x%02x)", indexedIfdName, tagId)

    if indexedIfdName == IfdName(IfdExif, 0) {
        if tagId == 0x9000 {
            // ExifVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0xa000 {
            // FlashpixVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0x9286 {
            // UserComment

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)

            unknownUc := TagUnknownType_9298_UserComment{
                EncodingType: TagUnknownType_9298_UserComment_Encoding_UNDEFINED,
                EncodingBytes: []byte{},
            }

            encoding := valueBytes[:8]
            for encodingIndex, encodingBytes := range TagUnknownType_9298_UserComment_Encodings {
                if bytes.Compare(encoding, encodingBytes) == 0 {
                    // If unknown, return the default rather than what we have
                    // because there will be a big list of NULs (which aren't
                    // functional) and this won't equal the default instance
                    // (above).
                    if encodingIndex == TagUnknownType_9298_UserComment_Encoding_UNDEFINED {
                        return unknownUc, nil
                    } else {
                        uc := TagUnknownType_9298_UserComment{
                            EncodingType: encodingIndex,
                            EncodingBytes: valueBytes[8:],
                        }

                        return uc, nil
                    }
                }
            }

            typeDecodeLogger.Warningf(nil, "User-comment encoding not valid. Returning 'unknown' type (the default).")
            return unknownUc, nil
        } else if tagId == 0x927c {
            // MakerNote
// TODO(dustin): !! This is the Wild Wild West. This very well might be a child IFD, but any and all OEM's define their own formats. If we're going to be writing changes and this is complete EXIF (which may not have the first eight bytes), it might be fine. However, if these are just IFDs they'll be relative to the main EXIF, this will invalidate the MakerNote data for IFDs and any other implementations that use offsets unless we can interpret them all. It be best to return to this later and just exclude this from being written for now, though means a loss of a wealth of image metadata.
//                  -> We can also just blindly try to interpret as an IFD and just validate that it's looks good (maybe it will even have a 'next ifd' pointer that we can validate is 0x0).

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)


// TODO(dustin): Doesn't work, but here as an example.
//             ie := NewIfdEnumerate(valueBytes, byteOrder)

// // TODO(dustin): !! Validate types (might have proprietary types, but it might be worth splitting the list between valid and not validate; maybe fail if a certain proportion are invalid, or maybe aren't less then a certain small integer)?
//             ii, err := ie.Collect(0x0)

//             for _, entry := range ii.RootIfd.Entries {
//                 fmt.Printf("ENTRY: 0x%02x %d\n", entry.TagId, entry.TagType)
//             }

            mn := TagUnknownType_927C_MakerNote{
                MakerNoteType: valueBytes[:20],

                // MakerNoteBytes has the whole length of bytes. There's always
                // the chance that the first 20 bytes includes actual data.
                MakerNoteBytes: valueBytes,
            }

            return mn, nil
        } else if tagId == 0x9101 {
            // ComponentsConfiguration

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)

            for configurationId, configurationBytes := range TagUnknownType_9101_ComponentsConfiguration_Configurations {
                if bytes.Compare(valueBytes, configurationBytes) == 0 {
                    cc := TagUnknownType_9101_ComponentsConfiguration{
                        ConfigurationId: configurationId,
                        ConfigurationBytes: valueBytes,
                    }

                    return cc, nil
                }
            }

            cc := TagUnknownType_9101_ComponentsConfiguration{
                ConfigurationId: TagUnknownType_9101_ComponentsConfiguration_OTHER,
                ConfigurationBytes: valueBytes,
            }

            return cc, nil
        }
    } else if indexedIfdName == IfdName(IfdGps, 0) {
        if tagId == 0x001c {
            // GPSAreaInformation

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0x001b {
            // GPSProcessingMethod

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        }
    } else if indexedIfdName == IfdName(IfdIop, 0) {
        if tagId == 0x0002 {
            // InteropVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        }
    }

// TODO(dustin): !! Still need to do:
//
// complex: 0xa302, 0xa20c, 0x8828
// long: 0xa301, 0xa300

    // 0xa40b is device-specific and unhandled.


    log.Panic(ErrUnhandledUnknownTypedTag)
    return nil, nil
}
