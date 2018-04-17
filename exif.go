package exif

import (
    "errors"
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    exifLogger = log.NewLogger("exif.exif")
    ErrNotExif = errors.New("not exif data")
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

func (e *Exif) Parse(data []byte, visitor TagVisitor) (err error) {
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
        exifLogger.Warningf(nil, "EXIF app-data header fixed-bytes should be 0x002a but are: [%v]", fixedBytes)
        return nil
    }

    firstIfdOffset := uint32(0)
    if byteOrder == binary.BigEndian {
        firstIfdOffset = binary.BigEndian.Uint32(data[10:14])
    } else {
        firstIfdOffset = binary.LittleEndian.Uint32(data[10:14])
    }

    ifd := NewIfd(data, byteOrder)

    err = ifd.Scan(IfdStandard, firstIfdOffset, visitor)
    log.PanicIf(err)

    return nil
}
