package exif

import (
    "fmt"
    "bytes"
    "io"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

const (
    BigEndianByteOrder = iota
    LittleEndianByteOrder = iota
)

var (
    ifdLogger = log.NewLogger("exifjpeg.ifd")
)

type IfdByteOrder int

func (ibo IfdByteOrder) IsBigEndian() bool {
    return ibo == BigEndianByteOrder
}

func (ibo IfdByteOrder) IsLittleEndian() bool {
    return ibo == LittleEndianByteOrder
}

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
func (ifd *Ifd) getUint32() (value uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ifd.read(ifd.buffer, &value)
    log.PanicIf(err)

    ifd.currentOffset += 4

    return value, nil
}

// parseCurrentIfd decodes the IFD block that we're currently sitting on the
// first byte of.
func (ifd *Ifd) parseCurrentIfd() (nextIfdOffset uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()


    tagCount, err := ifd.getUint16()
    log.PanicIf(err)

    fmt.Printf("IFD: TOTAL TAG COUNT=(%02x)\n", tagCount)

    t := NewTagIndex()

    for i := uint16(0); i < tagCount; i++ {
// TODO(dustin): !! 0x8769 tag-IDs are child IFDs.
        tagId, err := ifd.getUint16()
        log.PanicIf(err)

        it, err := t.GetWithTagId(tagId)
        if err != nil {
            if err == ErrTagNotFound {
                log.Panicf("tag (%04x) not known")
            } else {
                log.Panic(err)
            }
        }

        fmt.Printf("IFD: Tag (%d) ID=(%02x) NAME=[%s] IFD=[%s]\n", i, tagId, it.Name, it.Ifd)


        tagType, err := ifd.getUint16()
        log.PanicIf(err)

        fmt.Printf("IFD: Tag (%d) TYPE=(%d)\n", i, tagType)


        tagCount, err := ifd.getUint32()
        log.PanicIf(err)

        fmt.Printf("IFD: Tag (%d) COUNT=(%02x)\n", i, tagCount)


        valueOffset, err := ifd.getUint32()
        log.PanicIf(err)

        fmt.Printf("IFD: Tag (%d) VALUE-OFFSET=(%x)\n", i, valueOffset)

// Notes on the tag-value's value (we'll have to use this as a pointer if the type potentially requires more than four bytes):
//
// This tag records the offset from the start of the TIFF header to the position where the value itself is
// recorded. In cases where the value fits in 4 Bytes, the value itself is recorded. If the value is smaller
// than 4 Bytes, the value is stored in the 4-Byte area starting from the left, i.e., from the lower end of
// the byte offset area. For example, in big endian format, if the type is SHORT and the value is 1, it is
// recorded as 00010000.H

    }

    fmt.Printf("\n")

    nextIfdOffset, err = ifd.getUint32()
    log.PanicIf(err)

    fmt.Printf("IFD: NEXT-IFD-OFFSET=(%x)\n", nextIfdOffset)

    fmt.Printf("\n")

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

    fmt.Printf("IFD: Forwarding to IFD. TOP-OFFSET=(%d) IFD-OFFSET=(%d)\n", ifd.ifdTopOffset, ifdOffset)

    nextOffset := ifd.ifdTopOffset + ifdOffset

    // We're assuming the guarantee that the next IFD will follow the
    // current one. So, figure out how far it is from our current position.
    delta := nextOffset - ifd.currentOffset
    ifd.buffer.Next(int(delta))

    ifd.currentOffset = nextOffset

    return nil
}

type IfdVisitor func() error

// Scan enumerates the different EXIF blocks (called IFDs).
func (ifd *Ifd) Scan(v IfdVisitor, firstIfdOffset uint32) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ifd.forwardToIfd(firstIfdOffset)
    log.PanicIf(err)

    for {
        nextIfdOffset, err := ifd.parseCurrentIfd()
        log.PanicIf(err)

        if nextIfdOffset == 0 {
            break
        }

        err = ifd.forwardToIfd(nextIfdOffset)
        log.PanicIf(err)
    }

    return nil
}
