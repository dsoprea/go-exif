package exif

import (
    "os"
    "errors"
    "bytes"
    "fmt"

    "io/ioutil"
    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

const (
    // RootIfdExifOffset is the offset of the first IFD in the block of EXIF
    // data.
    RootIfdExifOffset = uint32(0x0008)

    // ExifAddressableAreaStart is the position that all offsets are relative
    // to. It's actually only about halfway through the header (technically
    // unallocatable space).
    ExifAddressableAreaStart = uint32(0x6)

    // ExifDefaultFirstIfdOffset is essentially the number of bytes in addition
    // to `ExifAddressableAreaStart` that you have to move in order to escape
    // the rest of the header and get to the earliest point where we can put
    // stuff (which has to be the first IFD). This is the size of the header
    // sequence containing the two-character byte-order, two-character fixed-
    // bytes, and the four bytes describing the first-IFD offset.
    ExifDefaultFirstIfdOffset = uint32(2 + 2 + 4)
)

var (
    exifLogger = log.NewLogger("exif.exif")

    ExifHeaderPrefixBytes = []byte("Exif\000\000")

    // EncodeDefaultByteOrder is the default byte-order for encoding operations.
    EncodeDefaultByteOrder = binary.BigEndian

    BigEndianBoBytes = [2]byte { 'M', 'M' }
    LittleEndianBoBytes = [2]byte { 'I', 'I' }

    ByteOrderLookup = map[[2]byte]binary.ByteOrder {
        BigEndianBoBytes: binary.BigEndian,
        LittleEndianBoBytes: binary.LittleEndian,
    }

    ByteOrderLookupR = map[binary.ByteOrder][2]byte {
        binary.BigEndian: BigEndianBoBytes,
        binary.LittleEndian: LittleEndianBoBytes,
    }

    ExifFixedBytes = [2]byte { 0x2a, 0x00 }
)

var (
    ErrNotExif = errors.New("not exif data")
    ErrExifHeaderError = errors.New("exif header error")
)


// SearchAndExtractExif returns a slice from the beginning of the EXIF data the
// end of the file (it's not practical to try and calculate where the data
// actually ends).
func SearchAndExtractExif(data []byte) (rawExif []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.Panic(err)
        }
    }()

    // Search for the beginning of the EXIF information. The EXIF is near the
    // beginning of our/most JPEGs, so this has a very low cost.

    foundAt := -1
    for i := 0; i < len(data); i++ {
        if _, err := ParseExifHeader(data[i:]); err == nil {
            foundAt = i
            break
        } else if log.Is(err, ErrNotExif) == false {
            log.Panic(err)
        }
    }

    if foundAt == -1 {
        log.Panicf("EXIF start not found")
    }

    return data[foundAt:], nil
}

// SearchFileAndExtractExif returns a slice from the beginning of the EXIF data
// to the end of the file (it's not practical to try and calculate where the
// data actually ends).
func SearchFileAndExtractExif(filepath string) (rawExif []byte, err error) {
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

    rawExif, err = SearchAndExtractExif(data)
    log.PanicIf(err)

    return rawExif, nil
}


type ExifHeader struct {
    ByteOrder binary.ByteOrder
    FirstIfdOffset uint32
}

func (eh ExifHeader) String() string {
    return fmt.Sprintf("ExifHeader<BYTE-ORDER=[%v] FIRST-IFD-OFFSET=(0x%02x)>", eh.ByteOrder, eh.FirstIfdOffset)
}

// ParseExifHeader parses the bytes at the very top of the header.
//
// This will panic with ErrNotExif on any data errors so that we can double as
// an EXIF-detection routine.
func ParseExifHeader(data []byte) (eh ExifHeader, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if bytes.Compare(data[:6], ExifHeaderPrefixBytes) != 0 {
        log.Panic(ErrNotExif)
    }

    // Good reference:
    //
    //      CIPA DC-008-2016; JEITA CP-3451D
    //      -> http://www.cipa.jp/std/documents/e/DC-008-Translation-2016-E.pdf

    byteOrderBytes := [2]byte { data[6], data[7] }

    byteOrder, found := ByteOrderLookup[byteOrderBytes]
    if found == false {
        exifLogger.Warningf(nil, "EXIF byte-order not recognized: [%v]", byteOrderBytes)
        log.Panic(ErrNotExif)
    }

    fixedBytes := [2]byte { data[8], data[9] }
    if fixedBytes != ExifFixedBytes {
        exifLogger.Warningf(nil, "EXIF header fixed-bytes should be 0x002a but are: [%v]", fixedBytes)
        log.Panic(ErrNotExif)
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

// Visit recursively invokes a callback for every tag.
func Visit(exifData []byte, visitor TagVisitor) (eh ExifHeader, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err = ParseExifHeader(exifData)
    log.PanicIf(err)

    ie := NewIfdEnumerate(exifData, eh.ByteOrder)

    err = ie.Scan(eh.FirstIfdOffset, visitor)
    log.PanicIf(err)

    return eh, nil
}

// Collect recursively builds a static structure of all IFDs and tags.
func Collect(exifData []byte) (eh ExifHeader, index IfdIndex, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err = ParseExifHeader(exifData)
    log.PanicIf(err)

    ie := NewIfdEnumerate(exifData, eh.ByteOrder)

    index, err = ie.Collect(eh.FirstIfdOffset)
    log.PanicIf(err)

    return eh, index, nil
}

// BuildExifHeader constructs the bytes that go in the very beginning.
func BuildExifHeader(byteOrder binary.ByteOrder, firstIfdOffset uint32) (headerBytes []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    b := new(bytes.Buffer)

    _, err = b.Write(ExifHeaderPrefixBytes)
    log.PanicIf(err)

// NOTE: This is the point in the data that all offsets are relative to.
    boBytes := ByteOrderLookupR[byteOrder]
    _, err = b.WriteString(string(boBytes[:]))
    log.PanicIf(err)

    _, err = b.Write(ExifFixedBytes[:])
    log.PanicIf(err)

    err = binary.Write(b, byteOrder, firstIfdOffset)
    log.PanicIf(err)

    return b.Bytes(), nil
}
