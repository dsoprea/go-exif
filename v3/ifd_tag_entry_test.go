package exif

import (
	"bytes"
	"testing"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestIfdTagEntry_RawBytes_Allocated(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	addressableBytes := data
	sb := rifs.NewSeekableBufferWithBytes(addressableBytes)

	ite := newIfdTagEntry(
		exifcommon.IfdStandardIfdIdentity,
		0x1,
		0,
		exifcommon.TypeByte,
		6,
		0,
		nil,
		sb,
		exifcommon.TestDefaultByteOrder)

	value, err := ite.GetRawBytes()
	log.PanicIf(err)

	if bytes.Compare(value, data) != 0 {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_RawBytes_Embedded(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)

			t.Fatalf("Test failure.")
		}
	}()

	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := newIfdTagEntry(
		exifcommon.IfdStandardIfdIdentity,
		0x1,
		0,
		exifcommon.TypeByte,
		4,
		0,
		data,
		nil,
		exifcommon.TestDefaultByteOrder)

	value, err := ite.GetRawBytes()
	log.PanicIf(err)

	if bytes.Compare(value, data) != 0 {
		t.Fatalf("Value not expected: %v != %v", value, data)
	}
}

func TestIfdTagEntry_String(t *testing.T) {
	ite := newIfdTagEntry(
		exifcommon.IfdStandardIfdIdentity,
		0x1,
		0,
		exifcommon.TypeByte,
		6,
		0,
		nil,
		nil,
		exifcommon.TestDefaultByteOrder)

	expected := "IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x0001) TAG-TYPE=[BYTE] UNIT-COUNT=(6)>"
	if ite.String() != expected {
		t.Fatalf("string representation not expected: [%s] != [%s]", ite.String(), expected)
	}
}
