package exif

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"encoding/binary"
	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

const (
	// ExifAddressableAreaStart is the absolute offset in the file that all
	// offsets are relative to.
	ExifAddressableAreaStart = uint32(0x0)

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

	ExifBigEndianSignature    = [4]byte{'M', 'M', 0x00, 0x2a}
	ExifLittleEndianSignature = [4]byte{'I', 'I', 0x2a, 0x00}
)

var (
	ErrNoExif          = errors.New("no exif data")
	ErrExifHeaderError = errors.New("exif header error")
)

// SearchAndExtractExif returns a slice from the beginning of the EXIF data to
// end of the file (it's not practical to try and calculate where the data
// actually ends; it needs to be formally parsed).
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
		} else if log.Is(err, ErrNoExif) == false {
			return nil, err
		}
	}

	if foundAt == -1 {
		return nil, ErrNoExif
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
	ByteOrder      binary.ByteOrder
	FirstIfdOffset uint32
}

func (eh ExifHeader) String() string {
	return fmt.Sprintf("ExifHeader<BYTE-ORDER=[%v] FIRST-IFD-OFFSET=(0x%02x)>", eh.ByteOrder, eh.FirstIfdOffset)
}

// ParseExifHeader parses the bytes at the very top of the header.
//
// This will panic with ErrNoExif on any data errors so that we can double as
// an EXIF-detection routine.
func ParseExifHeader(data []byte) (eh ExifHeader, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// Good reference:
	//
	//      CIPA DC-008-2016; JEITA CP-3451D
	//      -> http://www.cipa.jp/std/documents/e/DC-008-Translation-2016-E.pdf

	if len(data) < 8 {
		exifLogger.Warningf(nil, "Not enough data for EXIF header: (%d)", len(data))
		return eh, ErrNoExif
	}

	if bytes.Equal(data[:4], ExifBigEndianSignature[:]) == true {
		eh.ByteOrder = binary.BigEndian
	} else if bytes.Equal(data[:4], ExifLittleEndianSignature[:]) == true {
		eh.ByteOrder = binary.LittleEndian
	} else {
		return eh, ErrNoExif
	}

	eh.FirstIfdOffset = eh.ByteOrder.Uint32(data[4:8])

	return eh, nil
}

// Visit recursively invokes a callback for every tag.
func Visit(rootIfdName string, ifdMapping *IfdMapping, tagIndex *TagIndex, exifData []byte, visitor TagVisitorFn) (eh ExifHeader, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	eh, err = ParseExifHeader(exifData)
	log.PanicIf(err)

	ie := NewIfdEnumerate(ifdMapping, tagIndex, exifData, eh.ByteOrder)

	err = ie.Scan(rootIfdName, eh.FirstIfdOffset, visitor)
	log.PanicIf(err)

	return eh, nil
}

// Collect recursively builds a static structure of all IFDs and tags.
func Collect(ifdMapping *IfdMapping, tagIndex *TagIndex, exifData []byte) (eh ExifHeader, index IfdIndex, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	eh, err = ParseExifHeader(exifData)
	log.PanicIf(err)

	ie := NewIfdEnumerate(ifdMapping, tagIndex, exifData, eh.ByteOrder)

	index, err = ie.Collect(eh.FirstIfdOffset)
	log.PanicIf(err)

	return eh, index, nil
}

// BuildExifHeader constructs the bytes that go at the front of the stream.
func BuildExifHeader(byteOrder binary.ByteOrder, firstIfdOffset uint32) (headerBytes []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	b := new(bytes.Buffer)

	var signatureBytes []byte
	if byteOrder == binary.BigEndian {
		signatureBytes = ExifBigEndianSignature[:]
	} else {
		signatureBytes = ExifLittleEndianSignature[:]
	}

	_, err = b.Write(signatureBytes)
	log.PanicIf(err)

	err = binary.Write(b, byteOrder, firstIfdOffset)
	log.PanicIf(err)

	return b.Bytes(), nil
}
