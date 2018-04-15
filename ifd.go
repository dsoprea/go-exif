package exif

import (
    "bytes"
    "io"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    ifdLogger = log.NewLogger("exifjpeg.ifd")
)


type Ifd struct {
    data []byte
    buffer *bytes.Buffer
    byteOrder IfdByteOrder
    currentOffset uint32
    ifdTopOffset uint32
}

func NewIfd(data []byte, byteOrder IfdByteOrder) *Ifd {
    return &Ifd{
        data: data,
        buffer: bytes.NewBuffer(data),
        byteOrder: byteOrder,
        ifdTopOffset: 6,
    }
}

// read is a wrapper around the built-in reader that applies which endianness
// we are.
func (ifd *Ifd) read(r io.Reader, into interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if ifd.byteOrder.IsLittleEndian() == true {
        err := binary.Read(r, binary.LittleEndian, into)
        log.PanicIf(err)
    } else {
        err := binary.Read(r, binary.BigEndian, into)
        log.PanicIf(err)
    }

    return nil
}

// getUint16 reads a uint16 and advances both our current and our current
// accumulator (which allows us to know how far to seek to the beginning of the
// next IFD when it's time to jump).
func (ifd *Ifd) getUint16() (value uint16, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ifd.read(ifd.buffer, &value)
    log.PanicIf(err)

    ifd.currentOffset += 2

    return value, nil
}

// getUint32 reads a uint32 and advances both our current and our current
// accumulator (which allows us to know how far to seek to the beginning of the
// next IFD when it's time to jump).
func (ifd *Ifd) getUint32() (value uint32, raw []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    raw = make([]byte, 4)

    _, err = ifd.buffer.Read(raw)
    log.PanicIf(err)

    ifd.currentOffset += 4

    if ifd.byteOrder.IsBigEndian() {
        value = binary.BigEndian.Uint32(raw)
    } else {
        value = binary.LittleEndian.Uint32(raw)
    }

    return value, raw, nil
}

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
    UnitCount uint32
    ValueOffset uint32
    RawValueOffset []byte
    RawExif []byte
}


// TagVisitor is an optional callback that can get hit for every tag we parse
// through. `rawExif` is the byte array startign after the EXIF header (where
// the offsets of all IFDs and values are calculated from).
type TagVisitor func(tagId uint16, tagType TagType, valueContext ValueContext) (err error)

// parseCurrentIfd decodes the IFD block that we're currently sitting on the
// first byte of.
func (ifd *Ifd) parseCurrentIfd(visitor TagVisitor) (nextIfdOffset uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()


    tagCount, err := ifd.getUint16()
    log.PanicIf(err)

    ifdLogger.Debugf(nil, "Current IFD tag-count: (%d)", tagCount)

    for i := uint16(0); i < tagCount; i++ {

// TODO(dustin): !! 0x8769 tag-IDs are child IFDs. We need to be able to recurse.

        tagId, err := ifd.getUint16()
        log.PanicIf(err)

        tagType, err := ifd.getUint16()
        log.PanicIf(err)

        unitCount, _, err := ifd.getUint32()
        log.PanicIf(err)

        valueOffset, rawValueOffset, err := ifd.getUint32()
        log.PanicIf(err)

        if visitor != nil {
            tt := NewTagType(tagType, ifd.byteOrder)

            vc := ValueContext{
                UnitCount: unitCount,
                ValueOffset: valueOffset,
                RawValueOffset: rawValueOffset,
                RawExif: ifd.data[ifd.ifdTopOffset:],
            }

            err := visitor(tagId, tt, vc)
            log.PanicIf(err)
        }
    }

    nextIfdOffset, _, err = ifd.getUint32()
    log.PanicIf(err)

    ifdLogger.Debugf(nil, "Next IFD at offset: (%08x)", nextIfdOffset)

    return nextIfdOffset, nil
}

// forwardToIfd jumps to the beginning of an IFD block that starts on or after
// the current position.
func (ifd *Ifd) forwardToIfd(ifdOffset uint32) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ifdLogger.Debugf(nil, "Forwarding to IFD. TOP-OFFSET=(%d) IFD-OFFSET=(%d)", ifd.ifdTopOffset, ifdOffset)

    nextOffset := ifd.ifdTopOffset + ifdOffset

    // We're assuming the guarantee that the next IFD will follow the
    // current one. So, figure out how far it is from our current position.
    delta := nextOffset - ifd.currentOffset
    ifd.buffer.Next(int(delta))

    ifd.currentOffset = nextOffset

    return nil
}

// Scan enumerates the different EXIF blocks (called IFDs).
func (ifd *Ifd) Scan(visitor TagVisitor, firstIfdOffset uint32) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ifd.forwardToIfd(firstIfdOffset)
    log.PanicIf(err)

    for {
        nextIfdOffset, err := ifd.parseCurrentIfd(visitor)
        log.PanicIf(err)

        if nextIfdOffset == 0 {
            break
        }

        err = ifd.forwardToIfd(nextIfdOffset)
        log.PanicIf(err)
    }

    return nil
}
