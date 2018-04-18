package exif

import (
    "errors"
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    exifLogger = log.NewLogger("exif.exif")
)

var (
    ErrNotExif = errors.New("not exif data")
    ErrExifHeaderError = errors.New("exif header error")
)

type Exif struct {

}

func NewExif() *Exif {
    return new(Exif)
}

func (e *Exif) IsExif(data []byte) (ok bool) {
    if bytes.Compare(data[:6], []byte("Exif\000\000")) == 0 {
        return true
    }

    return false
}


type ExifHeader struct {
    ByteOrder binary.ByteOrder
    FirstIfdOffset uint32
}

func (e *Exif) ParseExifHeader(data []byte) (eh ExifHeader, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if e.IsExif(data) == false {
        log.Panic(ErrNotExif)
    }

    // Good reference:
    //
    //      CIPA DC-008-2016; JEITA CP-3451D
    //      -> http://www.cipa.jp/std/documents/e/DC-008-Translation-2016-E.pdf

    byteOrderSignature := data[6:8]
    var byteOrder binary.ByteOrder
    byteOrder = binary.BigEndian
    if string(byteOrderSignature) == "II" {
        byteOrder = binary.LittleEndian
    } else if string(byteOrderSignature) != "MM" {
        log.Panicf("byte-order not recognized: [%v]", byteOrderSignature)
    }

    fixedBytes := data[8:10]
    if fixedBytes[0] != 0x2a || fixedBytes[1] != 0x00 {
        exifLogger.Warningf(nil, "EXIF header fixed-bytes should be 0x002a but are: [%v]", fixedBytes)
        log.Panic(ErrExifHeaderError)
    }

    firstIfdOffset := uint32(0)
    if byteOrder == binary.BigEndian {
        firstIfdOffset = binary.BigEndian.Uint32(data[10:14])
    } else {
        firstIfdOffset = binary.LittleEndian.Uint32(data[10:14])
    }

    eh = ExifHeader{
        ByteOrder: byteOrder,
        FirstIfdOffset: firstIfdOffset,
    }

    return eh, nil
}

func (e *Exif) Visit(data []byte, visitor TagVisitor) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err := e.ParseExifHeader(data)
    log.PanicIf(err)

    ie := NewIfdEnumerate(data, eh.ByteOrder)

    err = ie.Scan(IfdStandard, eh.FirstIfdOffset, visitor)
    log.PanicIf(err)

    return nil
}
