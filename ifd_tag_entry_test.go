package exif

import (
	"bytes"
	"testing"

	log "github.com/dsoprea/go-logging"
)

func TestIfdTagEntry_ValueString_Allocated(t *testing.T) {
	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        IfdPathStandard,
	}

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	value, err := ite.ValueString(data, TestDefaultByteOrder)
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
		TagType:        TypeByte,
		UnitCount:      4,
		ValueOffset:    0,
		RawValueOffset: data,
		IfdPath:        IfdPathStandard,
	}

	value, err := ite.ValueString(nil, TestDefaultByteOrder)
	log.PanicIf(err)

	expected := "11 22 33 44"
	if value != expected {
		t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
	}
}

func TestIfdTagEntry_ValueString_Unknown(t *testing.T) {
	data := []uint8{'0', '2', '3', '0'}

	ite := IfdTagEntry{
		TagId:          0x9000,
		TagIndex:       0,
		TagType:        TypeUndefined,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        IfdPathStandardExif,
	}

	value, err := ite.ValueString(nil, TestDefaultByteOrder)
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
		TagType:        TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        IfdPathStandard,
	}

	data := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	value, err := ite.ValueBytes(data, TestDefaultByteOrder)
	log.PanicIf(err)

	if !bytes.Equal(value, data) {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_ValueBytes_Embedded(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        TypeByte,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        IfdPathStandard,
	}

	value, err := ite.ValueBytes(nil, TestDefaultByteOrder)
	log.PanicIf(err)

	if !bytes.Equal(value, data) {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_Value_Normal(t *testing.T) {
	data := []byte{0x11, 0x22, 0x33, 0x44}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        TypeByte,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        IfdPathStandard,
	}

	value, err := ite.Value(nil, TestDefaultByteOrder)
	log.PanicIf(err)

	if !bytes.Equal(value.([]byte), data) {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_Value_Unknown(t *testing.T) {
	data := []uint8{'0', '2', '3', '0'}

	ite := IfdTagEntry{
		TagId:          0x9000,
		TagIndex:       0,
		TagType:        TypeUndefined,
		UnitCount:      4,
		ValueOffset:    0x0,
		RawValueOffset: data,
		IfdPath:        IfdPathStandardExif,
	}

	value, err := ite.Value(nil, TestDefaultByteOrder)
	log.PanicIf(err)

	gs := value.(TagUnknownType_GeneralString)

	vb, err := gs.ValueBytes()
	log.PanicIf(err)

	if !bytes.Equal(vb, data) {
		t.Fatalf("Value not expected: [%s] != [%s]", value, data)
	}
}

func TestIfdTagEntry_String(t *testing.T) {
	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        TypeByte,
		UnitCount:      6,
		ValueOffset:    0x0,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        IfdPathStandard,
	}

	expected := "IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x0001) TAG-TYPE=[BYTE] UNIT-COUNT=(6)>"
	if ite.String() != expected {
		t.Fatalf("string representation not expected: [%s] != [%s]", ite.String(), expected)
	}
}

func TestIfdTagEntryValueResolver_ValueBytes(t *testing.T) {
	allocatedData := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}

	ite := IfdTagEntry{
		TagId:          0x1,
		TagIndex:       0,
		TagType:        TypeByte,
		UnitCount:      uint32(len(allocatedData)),
		ValueOffset:    0x8,
		RawValueOffset: []byte{0x0, 0x0, 0x0, 0x0},
		IfdPath:        IfdPathStandard,
	}

	headerBytes, err := BuildExifHeader(TestDefaultByteOrder, uint32(0))
	log.PanicIf(err)

	exifData := make([]byte, len(headerBytes)+len(allocatedData))
	copy(exifData[0:], headerBytes)
	copy(exifData[len(headerBytes):], allocatedData)

	itevr := NewIfdTagEntryValueResolver(exifData, TestDefaultByteOrder)

	value, err := itevr.ValueBytes(&ite)
	log.PanicIf(err)

	if !bytes.Equal(value, allocatedData) {
		t.Fatalf("bytes not expected: %v != %v", value, allocatedData)
	}
}
