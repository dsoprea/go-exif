package exif

import (
    "testing"
    "bytes"

    "github.com/dsoprea/go-logging"
)

func Test_IfdTagEntry_ValueString_Allocated(t *testing.T) {
    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 6,
        ValueOffset: 0x0,
        RawValueOffset: []byte { 0x0, 0x0, 0x0, 0x0 },
        Ii: RootIi,
    }

    data := []byte { 0x11, 0x22, 0x33, 0x44, 0x55, 0x66 }

    value, err := ite.ValueString(data, TestDefaultByteOrder)
    log.PanicIf(err)

    expected := "11 22 33 44 55 66"
    if value != expected {
        t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
    }
}

func Test_IfdTagEntry_ValueString_Embedded(t *testing.T) {
    data := []byte { 0x11, 0x22, 0x33, 0x44 }

    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 4,
        ValueOffset: 0,
        RawValueOffset: data,
        Ii: RootIi,
    }

    value, err := ite.ValueString(nil, TestDefaultByteOrder)
    log.PanicIf(err)

    expected := "11 22 33 44"
    if value != expected {
        t.Fatalf("Value not expected: [%s] != [%s]", value, expected)
    }
}

func Test_IfdTagEntry_ValueBytes_Allocated(t *testing.T) {
    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 6,
        ValueOffset: 0x0,
        RawValueOffset: []byte { 0x0, 0x0, 0x0, 0x0 },
        Ii: RootIi,
    }

    data := []byte { 0x11, 0x22, 0x33, 0x44, 0x55, 0x66 }

    value, err := ite.ValueBytes(data, TestDefaultByteOrder)
    log.PanicIf(err)

    if bytes.Compare(value, data) != 0 {
        t.Fatalf("Value not expected: [%s] != [%s]", value, data)
    }
}

func Test_IfdTagEntry_ValueBytes_Embedded(t *testing.T) {
    data := []byte { 0x11, 0x22, 0x33, 0x44 }

    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 4,
        ValueOffset: 0x0,
        RawValueOffset: data,
        Ii: RootIi,
    }

    value, err := ite.ValueBytes(nil, TestDefaultByteOrder)
    log.PanicIf(err)

    if bytes.Compare(value, data) != 0 {
        t.Fatalf("Value not expected: [%s] != [%s]", value, data)
    }
}

func Test_IfdTagEntry_Value_Normal(t *testing.T) {
    data := []byte { 0x11, 0x22, 0x33, 0x44 }

    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 4,
        ValueOffset: 0x0,
        RawValueOffset: data,
        Ii: RootIi,
    }

    value, err := ite.Value(nil, TestDefaultByteOrder)
    log.PanicIf(err)

    if bytes.Compare(value.([]byte), data) != 0 {
        t.Fatalf("Value not expected: [%s] != [%s]", value, data)
    }
}

func Test_IfdTagEntry_Value_Unknown(t *testing.T) {
    data := []uint8 { '0', '2', '3', '0' }

    ite := IfdTagEntry{
        TagId: 0x9000,
        TagIndex: 0,
        TagType: TypeUndefined,
        UnitCount: 4,
        ValueOffset: 0x0,
        RawValueOffset: data,
        Ii: ExifIi,
    }

    value, err := ite.Value(nil, TestDefaultByteOrder)
    log.PanicIf(err)

    gs := value.(TagUnknownType_GeneralString)

    vb, err := gs.ValueBytes()
    log.PanicIf(err)

    if bytes.Compare(vb, data) != 0 {
        t.Fatalf("Value not expected: [%s] != [%s]", value, data)
    }
}

func Test_IfdTagEntry_String(t *testing.T) {
    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 6,
        ValueOffset: 0x0,
        RawValueOffset: []byte { 0x0, 0x0, 0x0, 0x0 },
        Ii: RootIi,
    }

    expected := "IfdTagEntry<TAG-IFD=[] TAG-ID=(0x01) TAG-TYPE=[BYTE] UNIT-COUNT=(6)>"
    if ite.String() != expected {
        t.Fatalf("string representation not expected: [%s] != [%s]", ite.String(), expected)
    }
}

func Test_IfdTagEntryValueResolver_ValueBytes(t *testing.T) {
    ite := IfdTagEntry{
        TagId: 0x1,
        TagIndex: 0,
        TagType: TypeByte,
        UnitCount: 6,
        ValueOffset: 0x0,
        RawValueOffset: []byte { 0x0, 0x0, 0x0, 0x0 },
        Ii: RootIi,
    }

    exifData := make([]byte, ExifAddressableAreaStart + 6)
    copy(exifData, ExifHeaderPrefixBytes)

    allocatedData := []byte { 0x11, 0x22, 0x33, 0x44, 0x55, 0x66 }
    copy(exifData[ExifAddressableAreaStart:], allocatedData)

    itevr := NewIfdTagEntryValueResolver(exifData, TestDefaultByteOrder)

    value, err := itevr.ValueBytes(&ite)
    log.PanicIf(err)

    if bytes.Compare(value, allocatedData) != 0 {
        t.Fatalf("bytes not expected: %v != %v", value, allocatedData)
    }
}
