package exif

import (
    "fmt"
    "reflect"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    iteLogger = log.NewLogger("exif.ifd_tag_entry")
)

type IfdTagEntry struct {
    TagId uint16
    TagIndex int
    TagType uint16
    UnitCount uint32
    ValueOffset uint32
    RawValueOffset []byte

    // ChildIfdName is a name if this tag represents a child IFD.
    ChildIfdName string

    // IfdName is the IFD that this tag belongs to.
    IfdName string
}

func (ite IfdTagEntry) String() string {
    return fmt.Sprintf("IfdTagEntry<TAG-IFD=[%s] TAG-ID=(0x%02x) TAG-TYPE=[%s] UNIT-COUNT=(%d)>", ite.ChildIfdName, ite.TagId, TypeNames[ite.TagType], ite.UnitCount)
}

func (ite IfdTagEntry) ValueBytes(addressableData []byte, byteOrder binary.ByteOrder) (value []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if ite.TagType == TypeUndefined {
        valueContext := ValueContext{
            UnitCount: ite.UnitCount,
            ValueOffset: ite.ValueOffset,
            RawValueOffset: ite.RawValueOffset,
            AddressableData: addressableData,
        }

        value, err := UndefinedValue(ite.IfdName, ite.TagId, valueContext, byteOrder)
        log.PanicIf(err)

        switch value.(type) {
        case []byte:
            return value.([]byte), nil
        case string:
            return []byte(value.(string)), nil
        case UnknownTagValue:
            valueBytes, err := value.(UnknownTagValue).ValueBytes()
            log.PanicIf(err)

            return valueBytes, nil
        default:
// TODO(dustin): !! Finish translating the rest of the types (make reusable and replace into other similar implementations?)
            log.Panicf("can not produce bytes for unknown-type tag (0x%02x): [%s]", ite.TagId, reflect.TypeOf(value))
        }
    }

    originalType := NewTagType(ite.TagType, byteOrder)
    byteCount := uint32(originalType.Size()) * ite.UnitCount

    tt := NewTagType(TypeByte, byteOrder)

    if tt.ValueIsEmbedded(byteCount) == true {
        iteLogger.Debugf(nil, "Reading BYTE value (ITE; embedded).")

        // In this case, the bytes normally used for the offset are actually
        // data.
        value, err = tt.ParseBytes(ite.RawValueOffset, byteCount)
        log.PanicIf(err)
    } else {
        iteLogger.Debugf(nil, "Reading BYTE value (ITE; at offset).")

        value, err = tt.ParseBytes(addressableData[ite.ValueOffset:], byteCount)
        log.PanicIf(err)
    }

    return value, nil
}


// IfdTagEntryValueResolver instances know how to resolve the values for any
// tag for a particular EXIF block.
type IfdTagEntryValueResolver struct {
    addressableData []byte
    byteOrder binary.ByteOrder
}

func NewIfdTagEntryValueResolver(exifData []byte, byteOrder binary.ByteOrder) (itevr *IfdTagEntryValueResolver) {
    // Make it obvious what data we expect and when we don't get it.
    if IsExif(exifData) == false {
        log.Panicf("not exif data")
    }

    return &IfdTagEntryValueResolver{
        addressableData: exifData[ExifAddressableAreaStart:],
        byteOrder: byteOrder,
    }
}

func (itevr *IfdTagEntryValueResolver) ValueBytes(ite *IfdTagEntry) (value []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    value, err = ite.ValueBytes(itevr.addressableData, itevr.byteOrder)
    return value, err
}
