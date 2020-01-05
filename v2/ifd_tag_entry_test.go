package exif

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

func TestIfdTagEntry_ValueString_Allocated(t *testing.T) {
	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        exifcommon.IfdPathStandard,
	}

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	value, err := ite.ValueString(data, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expected := "11 22 33 44 55 66"
	if value != expected {
		t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
	}
}

func TestIfdTagEntry_ValueString_Embedded(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      4,
		ValueOffset:    0,
		RawValueOffset: data,
		IfdPath:        exifcommon.IfdPathStandard,
	}

	value, err := ite.ValueString(nil, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expected := "11 22 33 44"
	if value != expected {
		t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
	}
}

func TestIfdTagEntry_ValueString_Undefined(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)

			t.Fatalf("Test failure.")
		}
	}()

	data := []uint8{'0', '2', '3', '0'}

	ite := IfdTagEntry{
		TagId:          0x9000,
		TagIndex:       0,
		TagType:        exifcommon.TypeUndefined,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        exifcommon.IfdPathStandardExif,
	}

	value, err := ite.ValueString(nil, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	expected := "0230"
	if value != expected {
		t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
	}
}

func TestIfdTagEntry_ValueBytes_Allocated(t *testing.T) {
	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        exifcommon.IfdPathStandard,
	}

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	value, err := ite.ValueBytes(data, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Compare(value, data) != 0 {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_ValueBytes_Embedded(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        exifcommon.IfdPathStandard,
	}

	value, err := ite.ValueBytes(nil, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Compare(value, data) != 0 {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_Value_Normal(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        exifcommon.IfdPathStandard,
	}

	value, err := ite.Value(nil, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Compare(value.([]byte), data) != 0 {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_Value_Undefined(t *testing.T) {
	data := []uint8{'0', '2', '3', '0'}

	ite := IfdTagEntry{
		TagId:          0x9000,
		TagIndex:       0,
		TagType:        exifcommon.TypeUndefined,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        exifcommon.IfdPathStandardExif,
	}

	value, err := ite.Value(nil, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	s := value.(fmt.Stringer)
	recovered := []byte(s.String())

	if bytes.Compare(recovered, data) != 0 {
		t.Fatalf("Value not expected: [%s] != [%s]", recovered, data)
	}
}

func TestIfdTagEntry_String(t *testing.T) {
	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        exifcommon.TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        exifcommon.IfdPathStandard,
	}

	expected := "IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x0001) TAG-TYPE=[BYTE] UNIT-COUNT=(6)>"
	if ite.String() != expected {
		t.Fatalf("string representation not expected: [%s] != [%s]", ite.String(), expected)
	}
}
