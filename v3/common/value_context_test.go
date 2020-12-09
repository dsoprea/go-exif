package exifcommon

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
	"github.com/dsoprea/go-utility/v2/filesystem"
)

func TestNewValueContext(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}
	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeLong,
		TestDefaultByteOrder)

	if vc.ifdPath != "aa/bb" {
		t.Fatalf("ifdPath not correct: [%s]", vc.ifdPath)
	} else if vc.tagId != 0x1234 {
		t.Fatalf("tagId not correct: (0x%04x)", vc.tagId)
	} else if vc.unitCount != 11 {
		t.Fatalf("unitCount not correct: (%d)", vc.unitCount)
	} else if vc.valueOffset != 22 {
		t.Fatalf("valueOffset not correct: (%d)", vc.valueOffset)
	} else if bytes.Equal(vc.rawValueOffset, rawValueOffset) != true {
		t.Fatalf("rawValueOffset not correct: %v", vc.rawValueOffset)
	} else if vc.tagType != TypeLong {
		t.Fatalf("tagType not correct: (%d)", vc.tagType)
	} else if vc.byteOrder != TestDefaultByteOrder {
		t.Fatalf("byteOrder not correct: %v", vc.byteOrder)
	}

	recoveredBytes, err := ioutil.ReadAll(vc.AddressableData())
	log.PanicIf(err)

	if bytes.Equal(recoveredBytes, addressableData) != true {
		t.Fatalf("AddressableData() not correct: %v", recoveredBytes)
	}
}

func TestValueContext_SetUndefinedValueType__ErrorWhenNotUndefined(t *testing.T) {
	defer func() {
		if errRaw := recover(); errRaw != nil {
			err := errRaw.(error)
			if err.Error() != "can not set effective type for unknown-type tag because this is *not* an unknown-type tag" {
				t.Fatalf("Error not expected: [%s]", err.Error())
			}

			return
		}

		t.Fatalf("Expected error.")
	}()

	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeLong,
		TestDefaultByteOrder)

	vc.SetUndefinedValueType(TypeLong)
}

func TestValueContext_SetUndefinedValueType__Ok(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	vc.SetUndefinedValueType(TypeLong)

	if vc.tagType != TypeUndefined {
		t.Fatalf("Internal type not still 'undefined': (%d)", vc.tagType)
	} else if vc.undefinedValueTagType != TypeLong {
		t.Fatalf("Internal undefined-type not correct: (%d)", vc.undefinedValueTagType)
	} else if vc.effectiveValueType() != TypeLong {
		t.Fatalf("Effective tag not correct: (%d)", vc.effectiveValueType())
	}
}

func TestValueContext_effectiveValueType(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	vc.SetUndefinedValueType(TypeLong)

	if vc.tagType != TypeUndefined {
		t.Fatalf("Internal type not still 'undefined': (%d)", vc.tagType)
	} else if vc.undefinedValueTagType != TypeLong {
		t.Fatalf("Internal undefined-type not correct: (%d)", vc.undefinedValueTagType)
	} else if vc.effectiveValueType() != TypeLong {
		t.Fatalf("Effective tag not correct: (%d)", vc.effectiveValueType())
	}
}

func TestValueContext_UnitCount(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if vc.UnitCount() != 11 {
		t.Fatalf("UnitCount() not correct: (%d)", vc.UnitCount())
	}
}

func TestValueContext_ValueOffset(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if vc.ValueOffset() != 22 {
		t.Fatalf("ValueOffset() not correct: (%d)", vc.ValueOffset())
	}
}

func TestValueContext_RawValueOffset(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if bytes.Equal(vc.RawValueOffset(), rawValueOffset) != true {
		t.Fatalf("RawValueOffset() not correct: %v", vc.RawValueOffset())
	}
}

func TestValueContext_AddressableData(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	recoveredBytes, err := ioutil.ReadAll(vc.AddressableData())
	log.PanicIf(err)

	if bytes.Equal(recoveredBytes, addressableData) != true {
		t.Fatalf("AddressableData() not correct: %v", recoveredBytes)
	}
}

func TestValueContext_ByteOrder(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if vc.ByteOrder() != TestDefaultByteOrder {
		t.Fatalf("ByteOrder() not correct: %v", vc.ByteOrder())
	}
}

func TestValueContext_IfdPath(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if vc.IfdPath() != "aa/bb" {
		t.Fatalf("IfdPath() not correct: [%s]", vc.IfdPath())
	}
}

func TestValueContext_TagId(t *testing.T) {
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		11,
		22,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	if vc.TagId() != 0x1234 {
		t.Fatalf("TagId() not correct: (%d)", vc.TagId())
	}
}

func TestValueContext_isEmbedded__True(t *testing.T) {
	unitCount := uint32(4)
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		22,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	if vc.isEmbedded() != true {
		t.Fatalf("isEmbedded() not correct: %v", vc.isEmbedded())
	}
}

func TestValueContext_isEmbedded__False(t *testing.T) {
	unitCount := uint32(5)
	rawValueOffset := []byte{0, 0, 0, 22}

	addressableData := []byte{1, 2, 3, 4}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		22,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	if vc.isEmbedded() != false {
		t.Fatalf("isEmbedded() not correct: %v", vc.isEmbedded())
	}
}

func TestValueContext_readRawEncoded__IsEmbedded(t *testing.T) {
	unitCount := uint32(4)

	rawValueOffset := []byte{1, 2, 3, 4}

	// Ignored, in this case.
	valueOffset := uint32(0)

	addressableData := []byte{}
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	recovered, err := vc.readRawEncoded()
	log.PanicIf(err)

	if bytes.Equal(recovered, rawValueOffset) != true {
		t.Fatalf("Embedded value bytes not recovered correctly: %v", recovered)
	}
}

func TestValueContext_readRawEncoded__IsRelative(t *testing.T) {
	unitCount := uint32(5)

	// Ignored, in this case.
	rawValueOffset := []byte{0, 0, 0, 0}

	valueOffset := uint32(4)

	data := []byte{5, 6, 7, 8, 9}

	addressableData := []byte{1, 2, 3, 4}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	recovered, err := vc.readRawEncoded()
	log.PanicIf(err)

	if bytes.Equal(recovered, data) != true {
		t.Fatalf("Relative value bytes not recovered correctly: %v", recovered)
	}
}

func TestValueContext_Format__Byte(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "61 62 63 64 65 66 67 68" {
		t.Fatalf("Format not correct for bytes: [%s]", value)
	}
}

func TestValueContext_Format__Ascii(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeAscii,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "abcdefg" {
		t.Fatalf("Format not correct for ASCII: [%s]", value)
	}
}

func TestValueContext_Format__AsciiNoNul(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeAsciiNoNul,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "abcdefgh" {
		t.Fatalf("Format not correct for ASCII (no NUL): [%s]", value)
	}
}

func TestValueContext_Format__Short(t *testing.T) {
	unitCount := uint32(4)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 1, 0, 2, 0, 3, 0, 4}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeShort,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "[1 2 3 4]" {
		t.Fatalf("Format not correct for shorts: [%s]", value)
	}
}

func TestValueContext_Format__Long(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeLong,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "[1 2]" {
		t.Fatalf("Format not correct for longs: [%s]", value)
	}
}

func TestValueContext_Format__Rational(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeRational,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "[1/2 3/4]" {
		t.Fatalf("Format not correct for rationals: [%s]", value)
	}
}

func TestValueContext_Format__SignedLong(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeSignedLong,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "[1 2]" {
		t.Fatalf("Format not correct for signed-longs: [%s]", value)
	}
}

func TestValueContext_Format__SignedRational(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeSignedRational,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "[1/2 3/4]" {
		t.Fatalf("Format not correct for signed-rationals: [%s]", value)
	}
}

func TestValueContext_Format__Undefined__NoEffectiveType(t *testing.T) {
	defer func() {
		if errRaw := recover(); errRaw != nil {
			err := errRaw.(error)
			if err.Error() != "undefined-value type not set" {
				t.Fatalf("Error not expected: [%s]", err.Error())
			}

			return
		}

		t.Fatalf("Expected error.")
	}()

	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "61 62 63 64 65 66 67 68" {
		t.Fatalf("Format not correct for bytes: [%s]", value)
	}
}

func TestValueContext_Format__Undefined__HasEffectiveType(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeUndefined,
		TestDefaultByteOrder)

	vc.SetUndefinedValueType(TypeAscii)

	value, err := vc.Format()
	log.PanicIf(err)

	if value != "abcdefg" {
		t.Fatalf("Format not correct for undefined (with effective type of string): [%s]", value)
	}
}

func TestValueContext_FormatFirst__Bytes(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	value, err := vc.FormatFirst()
	log.PanicIf(err)

	if value != "61 62 63 64 65 66 67 68" {
		t.Fatalf("FormatFirst not correct for bytes: [%s]", value)
	}
}

func TestValueContext_FormatFirst__String(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeAscii,
		TestDefaultByteOrder)

	value, err := vc.FormatFirst()
	log.PanicIf(err)

	if value != "abcdefg" {
		t.Fatalf("FormatFirst not correct for ASCII: [%s]", value)
	}
}

func TestValueContext_FormatFirst__List(t *testing.T) {
	unitCount := uint32(4)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 1, 0, 2, 0, 3, 0, 4}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeShort,
		TestDefaultByteOrder)

	value, err := vc.FormatFirst()
	log.PanicIf(err)

	if value != "1..." {
		t.Fatalf("FormatFirst not correct for shorts: [%s]", value)
	}
}

func TestValueContext_ReadBytes(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeByte,
		TestDefaultByteOrder)

	value, err := vc.ReadBytes()
	log.PanicIf(err)

	if bytes.Equal(value, data) != true {
		t.Fatalf("ReadBytes not correct: %v", value)
	}
}

func TestValueContext_ReadAscii(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeAscii,
		TestDefaultByteOrder)

	value, err := vc.ReadAscii()
	log.PanicIf(err)

	if value != "abcdefg" {
		t.Fatalf("ReadAscii not correct: [%s]", value)
	}
}

func TestValueContext_ReadAsciiNoNul(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeAsciiNoNul,
		TestDefaultByteOrder)

	value, err := vc.ReadAsciiNoNul()
	log.PanicIf(err)

	if value != "abcdefgh" {
		t.Fatalf("ReadAsciiNoNul not correct: [%s]", value)
	}
}

func TestValueContext_ReadShorts(t *testing.T) {
	unitCount := uint32(4)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 1, 0, 2, 0, 3, 0, 4}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeShort,
		TestDefaultByteOrder)

	value, err := vc.ReadShorts()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint16{1, 2, 3, 4}) != true {
		t.Fatalf("ReadShorts not correct: %v", value)
	}
}

func TestValueContext_ReadLongs(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeLong,
		TestDefaultByteOrder)

	value, err := vc.ReadLongs()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint32{1, 2}) != true {
		t.Fatalf("ReadLongs not correct: %v", value)
	}
}

func TestValueContext_ReadFloats(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0x40, 0x49, 0x0f, 0xdb, 0x40, 0x2d, 0xf8, 0x54}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeFloat,
		TestDefaultByteOrder)

	value, err := vc.ReadFloats()
	log.PanicIf(err)

	expectedResult := []float32{3.14159265, 2.71828182}
	for i, v := range value {
		if v < expectedResult[i] || v >= math.Nextafter32(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("ReadFloats expecting %v, received %v", expectedResult[i], v)
		}
	}
}
func TestValueContext_ReadDoubles(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0x40, 0x09, 0x21, 0xfb, 0x53, 0xc8, 0xd4, 0xf1,
		0x40, 0x05, 0xbf, 0x0a, 0x89, 0xf1, 0xb0, 0xdd}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeDouble,
		TestDefaultByteOrder)

	value, err := vc.ReadDoubles()
	log.PanicIf(err)

	expectedResult := []float64{3.14159265, 2.71828182}
	for i, v := range value {
		if v < expectedResult[i] || v >= math.Nextafter(expectedResult[i], expectedResult[i]+1) {
			t.Fatalf("ReadDoubles expecting %v, received %v", expectedResult[i], v)
		}
	}
}

func TestValueContext_ReadRationals(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeRational,
		TestDefaultByteOrder)

	value, err := vc.ReadRationals()
	log.PanicIf(err)

	expected := []Rational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("ReadRationals not correct: %v", value)
	}
}

func TestValueContext_ReadSignedLongs(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext(
		"aa/bb",
		0x1234,
		unitCount,
		valueOffset,
		rawValueOffset,
		sb,
		TypeSignedLong,
		TestDefaultByteOrder)

	value, err := vc.ReadSignedLongs()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []int32{1, 2}) != true {
		t.Fatalf("ReadSignedLongs not correct: %v", value)
	}
}

func TestValueContext_ReadSignedRationals(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeSignedRational, TestDefaultByteOrder)

	value, err := vc.ReadSignedRationals()
	log.PanicIf(err)

	expected := []SignedRational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("ReadSignedRationals not correct: %v", value)
	}
}

func TestValueContext_Values__Byte(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeByte, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, data) != true {
		t.Fatalf("Values not correct (bytes): %v", value)
	}
}

func TestValueContext_Values__Ascii(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 0}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeAscii, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, "abcdefg") != true {
		t.Fatalf("Values not correct (ASCII): [%s]", value)
	}
}

func TestValueContext_Values__AsciiNoNul(t *testing.T) {
	unitCount := uint32(8)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeAsciiNoNul, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, "abcdefgh") != true {
		t.Fatalf("Values not correct (ASCII no-NUL): [%s]", value)
	}
}

func TestValueContext_Values__Short(t *testing.T) {
	unitCount := uint32(4)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 1, 0, 2, 0, 3, 0, 4}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeShort, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint16{1, 2, 3, 4}) != true {
		t.Fatalf("Values not correct (shorts): %v", value)
	}
}

func TestValueContext_Values__Long(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeLong, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []uint32{1, 2}) != true {
		t.Fatalf("Values not correct (longs): %v", value)
	}
}

func TestValueContext_Values__Rational(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeRational, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	expected := []Rational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Values not correct (rationals): %v", value)
	}
}

func TestValueContext_Values__SignedLong(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{0, 0, 0, 1, 0, 0, 0, 2}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeSignedLong, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	if reflect.DeepEqual(value, []int32{1, 2}) != true {
		t.Fatalf("Values not correct (signed longs): %v", value)
	}
}

func TestValueContext_Values__SignedRational(t *testing.T) {
	unitCount := uint32(2)

	rawValueOffset := []byte{0, 0, 0, 4}
	valueOffset := uint32(4)

	data := []byte{
		0, 0, 0, 1, 0, 0, 0, 2,
		0, 0, 0, 3, 0, 0, 0, 4,
	}

	addressableData := []byte{0, 0, 0, 0}
	addressableData = append(addressableData, data...)
	sb := rifs.NewSeekableBufferWithBytes(addressableData)

	vc := NewValueContext("aa/bb", 0x1234, unitCount, valueOffset, rawValueOffset, sb, TypeSignedRational, TestDefaultByteOrder)

	value, err := vc.Values()
	log.PanicIf(err)

	expected := []SignedRational{
		{Numerator: 1, Denominator: 2},
		{Numerator: 3, Denominator: 4},
	}

	if reflect.DeepEqual(value, expected) != true {
		t.Fatalf("Values not correct (signed rationals): %v", value)
	}
}
