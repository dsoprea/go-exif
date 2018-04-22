package exif

import (
    "os"
    "errors"
    "bytes"

    "io/ioutil"
    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

const (
    // RootIfdExifOffset is the offset of the first IFD in the block of EXIF
    // data.
    RootIfdExifOffset = uint32(0x0008)

    // ExifAddressableAreaStart is the start of where all offsets are relative to.
    ExifAddressableAreaStart = uint32(0x6)
)

var (
    exifLogger = log.NewLogger("exif.exif")
)

var (
    ErrNotExif = errors.New("not exif data")
    ErrExifHeaderError = errors.New("exif header error")
)


func IsExif(data []byte) (ok bool) {
    if bytes.Compare(data[:6], []byte("Exif\000\000")) == 0 {
        return true
    }

    return false
}


type Exif struct {

}

func NewExif() *Exif {
    return new(Exif)
}

func (e *Exif) SearchAndExtractExif(filepath string) (rawExif []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.Panic(err)
        }
    }()

    // Open the file.

    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    data, err := ioutil.ReadAll(f)
    log.PanicIf(err)

    // Search for the beginning of the EXIF information. The EXIF is near the
    // beginning of our/most JPEGs, so this has a very low cost.

    foundAt := -1
    for i := 0; i < len(data); i++ {
        if IsExif(data[i:i + 6]) == true {
            foundAt = i
            break
        }
    }

    if foundAt == -1 {
        log.Panicf("EXIF start not found")
    }

    return data[foundAt:], nil
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

    if IsExif(data) == false {
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

func (e *Exif) Visit(exifData []byte, visitor TagVisitor) (eh ExifHeader, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err = e.ParseExifHeader(exifData)
    log.PanicIf(err)

    ie := NewIfdEnumerate(exifData, eh.ByteOrder)

    err = ie.Scan(IfdStandard, eh.FirstIfdOffset, visitor)
    log.PanicIf(err)

    return eh, nil
}

func (e *Exif) Collect(exifData []byte) (eh ExifHeader, index IfdIndex, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err = e.ParseExifHeader(exifData)
    log.PanicIf(err)

    ie := NewIfdEnumerate(exifData, eh.ByteOrder)

    index, err = ie.Collect(eh.FirstIfdOffset)
    log.PanicIf(err)

    return eh, index, nil
}
