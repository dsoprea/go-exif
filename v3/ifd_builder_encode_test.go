package exif

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func Test_ByteWriter_writeAsBytes_uint8(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.writeAsBytes(uint8(0x12))
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x12}) != 0 {
		t.Fatalf("uint8 not encoded correctly.")
	}
}

func Test_ByteWriter_writeAsBytes_uint16(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.writeAsBytes(uint16(0x1234))
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x12, 0x34}) != 0 {
		t.Fatalf("uint16 not encoded correctly.")
	}
}

func Test_ByteWriter_writeAsBytes_uint32(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.writeAsBytes(uint32(0x12345678))
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x12, 0x34, 0x56, 0x78}) != 0 {
		t.Fatalf("uint32 not encoded correctly.")
	}
}

func Test_ByteWriter_WriteUint16(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.WriteUint16(uint16(0x1234))
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x12, 0x34}) != 0 {
		t.Fatalf("uint16 not encoded correctly (as bytes).")
	}
}

func Test_ByteWriter_WriteUint32(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.WriteUint32(uint32(0x12345678))
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x12, 0x34, 0x56, 0x78}) != 0 {
		t.Fatalf("uint32 not encoded correctly (as bytes).")
	}
}

func Test_ByteWriter_WriteFourBytes(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.WriteFourBytes([]byte{0x11, 0x22, 0x33, 0x44})
	log.PanicIf(err)

	if bytes.Compare(b.Bytes(), []byte{0x11, 0x22, 0x33, 0x44}) != 0 {
		t.Fatalf("four-bytes not encoded correctly.")
	}
}

func Test_ByteWriter_WriteFourBytes_TooMany(t *testing.T) {
	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	err := bw.WriteFourBytes([]byte{0x11, 0x22, 0x33, 0x44, 0x55})
	if err == nil {
		t.Fatalf("expected error for not exactly four-bytes")
	} else if err.Error() != "value is not four-bytes: (5)" {
		t.Fatalf("wrong error for not exactly four bytes: %v", err)
	}
}

func Test_IfdDataAllocator_Allocate_InitialOffset1(t *testing.T) {
	addressableOffset := uint32(0)
	ida := newIfdDataAllocator(addressableOffset)

	if ida.NextOffset() != addressableOffset {
		t.Fatalf("initial offset not correct: (%d) != (%d)", ida.NextOffset(), addressableOffset)
	} else if len(ida.Bytes()) != 0 {
		t.Fatalf("initial buffer not empty")
	}

	data := []byte{0x1, 0x2, 0x3}
	offset, err := ida.Allocate(data)
	log.PanicIf(err)

	expected := uint32(addressableOffset + 0)
	if offset != expected {
		t.Fatalf("offset not bumped correctly (2): (%d) != (%d)", offset, expected)
	} else if ida.NextOffset() != offset+uint32(3) {
		t.Fatalf("position counter not advanced properly")
	} else if bytes.Compare(ida.Bytes(), []byte{0x1, 0x2, 0x3}) != 0 {
		t.Fatalf("buffer not correct after write (1)")
	}

	data = []byte{0x4, 0x5, 0x6}
	offset, err = ida.Allocate(data)
	log.PanicIf(err)

	expected = uint32(addressableOffset + 3)
	if offset != expected {
		t.Fatalf("offset not bumped correctly (3): (%d) != (%d)", offset, expected)
	} else if ida.NextOffset() != offset+uint32(3) {
		t.Fatalf("position counter not advanced properly")
	} else if bytes.Compare(ida.Bytes(), []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6}) != 0 {
		t.Fatalf("buffer not correct after write (2)")
	}
}

func Test_IfdDataAllocator_Allocate_InitialOffset2(t *testing.T) {
	addressableOffset := uint32(10)
	ida := newIfdDataAllocator(addressableOffset)

	if ida.NextOffset() != addressableOffset {
		t.Fatalf("initial offset not correct: (%d) != (%d)", ida.NextOffset(), addressableOffset)
	} else if len(ida.Bytes()) != 0 {
		t.Fatalf("initial buffer not empty")
	}

	data := []byte{0x1, 0x2, 0x3}
	offset, err := ida.Allocate(data)
	log.PanicIf(err)

	expected := uint32(addressableOffset + 0)
	if offset != expected {
		t.Fatalf("offset not bumped correctly (2): (%d) != (%d)", offset, expected)
	} else if ida.NextOffset() != offset+uint32(3) {
		t.Fatalf("position counter not advanced properly")
	} else if bytes.Compare(ida.Bytes(), []byte{0x1, 0x2, 0x3}) != 0 {
		t.Fatalf("buffer not correct after write (1)")
	}

	data = []byte{0x4, 0x5, 0x6}
	offset, err = ida.Allocate(data)
	log.PanicIf(err)

	expected = uint32(addressableOffset + 3)
	if offset != expected {
		t.Fatalf("offset not bumped correctly (3): (%d) != (%d)", offset, expected)
	} else if ida.NextOffset() != offset+uint32(3) {
		t.Fatalf("position counter not advanced properly")
	} else if bytes.Compare(ida.Bytes(), []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6}) != 0 {
		t.Fatalf("buffer not correct after write (2)")
	}
}

func Test_IfdByteEncoder__Arithmetic(t *testing.T) {
	ibe := NewIfdByteEncoder()

	if (ibe.TableSize(1) - ibe.TableSize(0)) != IfdTagEntrySize {
		t.Fatalf("table-size/entry-size not consistent (1)")
	} else if (ibe.TableSize(11) - ibe.TableSize(10)) != IfdTagEntrySize {
		t.Fatalf("table-size/entry-size not consistent (2)")
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_embedded1(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	ibe := NewIfdByteEncoder()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdGpsInfoStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	it, err := ti.Get(ib.IfdIdentity(), uint16(0x0000))
	log.PanicIf(err)

	bt := NewStandardBuilderTag(exifcommon.IfdGpsInfoStandardIfdIdentity.UnindexedString(), it, exifcommon.TestDefaultByteOrder, []uint8{uint8(0x12)})

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
	log.PanicIf(err)

	if childIfdBlock != nil {
		t.Fatalf("no child-IFDs were expected to be allocated")
	} else if bytes.Compare(b.Bytes(), []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x12, 0x00, 0x00, 0x00}) != 0 {
		t.Fatalf("encoded tag-entry bytes not correct")
	} else if ida.NextOffset() != addressableOffset {
		t.Fatalf("allocation was done but not expected")
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_embedded2(t *testing.T) {
	ibe := NewIfdByteEncoder()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdGpsInfoStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	it, err := ti.Get(ib.IfdIdentity(), uint16(0x0000))
	log.PanicIf(err)

	bt := NewStandardBuilderTag(exifcommon.IfdGpsInfoStandardIfdIdentity.UnindexedString(), it, exifcommon.TestDefaultByteOrder, []uint8{uint8(0x12), uint8(0x34), uint8(0x56), uint8(0x78)})

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
	log.PanicIf(err)

	if childIfdBlock != nil {
		t.Fatalf("no child-IFDs were expected to be allocated")
	} else if bytes.Compare(b.Bytes(), []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x12, 0x34, 0x56, 0x78}) != 0 {
		t.Fatalf("encoded tag-entry bytes not correct")
	} else if ida.NextOffset() != addressableOffset {
		t.Fatalf("allocation was done but not expected")
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_allocated(t *testing.T) {
	ibe := NewIfdByteEncoder()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdGpsInfoStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	it, err := ti.Get(ib.IfdIdentity(), uint16(0x0000))
	log.PanicIf(err)

	bt := NewStandardBuilderTag(exifcommon.IfdGpsInfoStandardIfdIdentity.UnindexedString(), it, exifcommon.TestDefaultByteOrder, []uint8{uint8(0x12), uint8(0x34), uint8(0x56), uint8(0x78), uint8(0x9a)})

	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
	log.PanicIf(err)

	if childIfdBlock != nil {
		t.Fatalf("no child-IFDs were expected to be allocated (1)")
	} else if bytes.Compare(b.Bytes(), []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x34}) != 0 {
		t.Fatalf("encoded tag-entry bytes not correct (1)")
	} else if ida.NextOffset() != addressableOffset+uint32(5) {
		t.Fatalf("allocation offset not expected (1)")
	} else if bytes.Compare(ida.Bytes(), []byte{0x12, 0x34, 0x56, 0x78, 0x9A}) != 0 {
		t.Fatalf("allocated data not correct (1)")
	}

	// Test that another allocation encodes to the new offset.

	bt = NewStandardBuilderTag(exifcommon.IfdGpsInfoStandardIfdIdentity.UnindexedString(), it, exifcommon.TestDefaultByteOrder, []uint8{uint8(0xbc), uint8(0xde), uint8(0xf0), uint8(0x12), uint8(0x34)})

	childIfdBlock, err = ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
	log.PanicIf(err)

	if childIfdBlock != nil {
		t.Fatalf("no child-IFDs were expected to be allocated (2)")
	} else if bytes.Compare(b.Bytes(), []byte{
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x34, // Tag 1
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x39, // Tag 2
	}) != 0 {
		t.Fatalf("encoded tag-entry bytes not correct (2)")
	} else if ida.NextOffset() != addressableOffset+uint32(10) {
		t.Fatalf("allocation offset not expected (2)")
	} else if bytes.Compare(ida.Bytes(), []byte{
		0x12, 0x34, 0x56, 0x78, 0x9A,
		0xbc, 0xde, 0xf0, 0x12, 0x34,
	}) != 0 {
		t.Fatalf("allocated data not correct (2)")
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_childIfd__withoutAllocate(t *testing.T) {
	ibe := NewIfdByteEncoder()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	childIb := NewIfdBuilder(im, ti, exifcommon.IfdExifStandardIfdIdentity, exifcommon.TestDefaultByteOrder)
	tagValue := NewIfdBuilderTagValueFromIfdBuilder(childIb)
	bt := NewChildIfdBuilderTag(exifcommon.IfdStandardIfdIdentity.UnindexedString(), exifcommon.IfdExifStandardIfdIdentity.TagId(), tagValue)

	nextIfdOffsetToWrite := uint32(0)
	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, nextIfdOffsetToWrite)
	log.PanicIf(err)

	if childIfdBlock != nil {
		t.Fatalf("no child-IFDs were expected to be allocated")
	} else if bytes.Compare(b.Bytes(), []byte{0x87, 0x69, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}) != 0 {
		t.Fatalf("encoded tag-entry with child-IFD not correct")
	} else if ida.NextOffset() != addressableOffset {
		t.Fatalf("allocation offset not expected")
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_childIfd__withAllocate(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	// Create a child IFD (represented by an IB instance) that we can allocate
	// space for and then attach to a tag (which would normally be an entry,
	// then, in a higher IFD).

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	childIb := NewIfdBuilder(im, ti, exifcommon.IfdExifStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	childIbTestTag := &BuilderTag{
		ifdPath: exifcommon.IfdExifStandardIfdIdentity.UnindexedString(),
		tagId:   0x8822,
		typeId:  exifcommon.TypeShort,
		value:   NewIfdBuilderTagValueFromBytes([]byte{0x12, 0x34}),
	}

	childIb.Add(childIbTestTag)

	// Formally compose the tag that refers to it.

	tagValue := NewIfdBuilderTagValueFromIfdBuilder(childIb)
	bt := NewChildIfdBuilderTag(exifcommon.IfdStandardIfdIdentity.UnindexedString(), exifcommon.IfdExifStandardIfdIdentity.TagId(), tagValue)

	// Encode the tag. Since we've actually provided an offset at which we can
	// allocate data, the child-IFD will automatically be encoded, allocated,
	// and installed into the allocated-data block (which will follow the IFD
	// block/table in the file).

	ibe := NewIfdByteEncoder()

	ib := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	// addressableOffset is the offset of where large data can be allocated
	// (which follows the IFD table/block). Large, in that it can't be stored
	// in the table itself. Just used for arithmetic. This is just where the
	// data for the current IFD can be written. It's not absolute for the EXIF
	// data in general.
	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	// This is the offset of where the next IFD can be written in the EXIF byte
	// stream. Just used for arithmetic.
	nextIfdOffsetToWrite := uint32(2000)

	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, nextIfdOffsetToWrite)
	log.PanicIf(err)

	if ida.NextOffset() != addressableOffset {
		t.Fatalf("IDA offset changed but no allocations where expected: (0x%02x)", ida.NextOffset())
	}

	tagBytes := b.Bytes()

	if len(tagBytes) != 12 {
		t.Fatalf("Tag not encoded to the right number of bytes: (%d)", len(tagBytes))
	} else if len(childIfdBlock) != 18 {
		t.Fatalf("Child IFD is not the right size: (%d)", len(childIfdBlock))
	}

	iteV, err := parseOneTag(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder, tagBytes)
	log.PanicIf(err)

	if iteV.TagId() != exifcommon.IfdExifStandardIfdIdentity.TagId() {
		t.Fatalf("IFD first tag-ID not correct: (0x%02x)", iteV.TagId())
	} else if iteV.tagIndex != 0 {
		t.Fatalf("IFD first tag index not correct: (%d)", iteV.tagIndex)
	} else if iteV.TagType() != exifcommon.TypeLong {
		t.Fatalf("IFD first tag type not correct: (%d)", iteV.TagType())
	} else if iteV.UnitCount() != 1 {
		t.Fatalf("IFD first tag unit-count not correct: (%d)", iteV.UnitCount())
	} else if iteV.getValueOffset() != nextIfdOffsetToWrite {
		t.Fatalf("IFD's child-IFD offset (as offset) is not correct: (%d) != (%d)", iteV.getValueOffset(), nextIfdOffsetToWrite)
	} else if iteV.ChildIfdPath() != exifcommon.IfdExifStandardIfdIdentity.UnindexedString() {
		t.Fatalf("IFD first tag IFD-name name not correct: [%s]", iteV.ChildIfdPath())
	} else if iteV.IfdPath() != exifcommon.IfdStandardIfdIdentity.UnindexedString() {
		t.Fatalf("IFD first tag parent IFD not correct: %v", iteV.IfdPath())
	}

	// Validate the child's raw IFD bytes.

	childNextIfdOffset, childEntries, err := parseOneIfd(im, ti, exifcommon.IfdExifStandardIfdIdentity, exifcommon.TestDefaultByteOrder, childIfdBlock, nil)
	log.PanicIf(err)

	if childNextIfdOffset != uint32(0) {
		t.Fatalf("Child IFD: Next IFD offset should be (0): (0x%08x)", childNextIfdOffset)
	} else if len(childEntries) != 1 {
		t.Fatalf("Child IFD: Expected exactly one entry: (%d)", len(childEntries))
	}

	ite := childEntries[0]

	if ite.TagId() != 0x8822 {
		t.Fatalf("Child IFD first tag-ID not correct: (0x%02x)", ite.TagId())
	} else if ite.tagIndex != 0 {
		t.Fatalf("Child IFD first tag index not correct: (%d)", ite.tagIndex)
	} else if ite.TagType() != exifcommon.TypeShort {
		t.Fatalf("Child IFD first tag type not correct: (%d)", ite.TagType())
	} else if ite.UnitCount() != 1 {
		t.Fatalf("Child IFD first tag unit-count not correct: (%d)", ite.UnitCount())
	} else if ite.ChildIfdPath() != "" {
		t.Fatalf("Child IFD first tag IFD-name name not empty: [%s]", ite.ChildIfdPath())
	} else if ite.IfdPath() != exifcommon.IfdExifStandardIfdIdentity.UnindexedString() {
		t.Fatalf("Child IFD first tag parent IFD not correct: %v", ite.IfdPath())
	}
}

func Test_IfdByteEncoder_encodeTagToBytes_simpleTag_allocate(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	// Encode the tag. Since we've actually provided an offset at which we can
	// allocate data, the child-IFD will automatically be encoded, allocated,
	// and installed into the allocated-data block (which will follow the IFD
	// block/table in the file).

	ibe := NewIfdByteEncoder()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	it, err := ib.tagIndex.Get(ib.IfdIdentity(), uint16(0x000b))
	log.PanicIf(err)

	valueString := "testvalue"
	bt := NewStandardBuilderTag(exifcommon.IfdStandardIfdIdentity.UnindexedString(), it, exifcommon.TestDefaultByteOrder, valueString)

	b := new(bytes.Buffer)
	bw := NewByteWriter(b, exifcommon.TestDefaultByteOrder)

	// addressableOffset is the offset of where large data can be allocated
	// (which follows the IFD table/block). Large, in that it can't be stored
	// in the table itself. Just used for arithmetic. This is just where the
	// data for the current IFD can be written. It's not absolute for the EXIF
	// data in general.
	addressableOffset := uint32(0x1234)
	ida := newIfdDataAllocator(addressableOffset)

	childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
	log.PanicIf(err)

	if ida.NextOffset() == addressableOffset {
		t.Fatalf("IDA offset did not change even though there should've been an allocation.")
	}

	tagBytes := b.Bytes()

	if len(tagBytes) != 12 {
		t.Fatalf("Tag not encoded to the right number of bytes: (%d)", len(tagBytes))
	} else if len(childIfdBlock) != 0 {
		t.Fatalf("Child IFD not have been allocated.")
	}

	ite, err := parseOneTag(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder, tagBytes)
	log.PanicIf(err)

	if ite.TagId() != 0x000b {
		t.Fatalf("Tag-ID not correct: (0x%02x)", ite.TagId())
	} else if ite.tagIndex != 0 {
		t.Fatalf("Tag index not correct: (%d)", ite.tagIndex)
	} else if ite.TagType() != exifcommon.TypeAscii {
		t.Fatalf("Tag type not correct: (%d)", ite.TagType())
	} else if ite.UnitCount() != (uint32(len(valueString) + 1)) {
		t.Fatalf("Tag unit-count not correct: (%d)", ite.UnitCount())
	} else if ite.ChildIfdPath() != "" {
		t.Fatalf("Tag's IFD-name should be empty: [%s]", ite.ChildIfdPath())
	} else if ite.IfdPath() != exifcommon.IfdStandardIfdIdentity.UnindexedString() {
		t.Fatalf("Tag's parent IFD is not correct: %v", ite.IfdPath())
	}

	expectedBuffer := bytes.NewBufferString(valueString)
	expectedBuffer.Write([]byte{0x0})
	expectedBytes := expectedBuffer.Bytes()

	allocatedBytes := ida.Bytes()

	if bytes.Compare(allocatedBytes, expectedBytes) != 0 {
		t.Fatalf("Allocated bytes not correct: %v != %v", allocatedBytes, expectedBytes)
	}
}

func Test_IfdByteEncoder_encodeIfdToBytes_simple(t *testing.T) {
	ib := getExifSimpleTestIb()

	// Write the byte stream.

	ibe := NewIfdByteEncoder()

	// addressableOffset is the offset of where large data can be allocated
	// (which follows the IFD table/block). Large, in that it can't be stored
	// in the table itself. Just used for arithmetic. This is just where the
	// data for the current IFD can be written. It's not absolute for the EXIF
	// data in general.
	addressableOffset := uint32(0x1234)

	tableAndAllocated, tableSize, allocatedDataSize, childIfdSizes, err := ibe.encodeIfdToBytes(ib, addressableOffset, uint32(0), false)
	log.PanicIf(err)

	expectedTableSize := ibe.TableSize(4)
	if tableSize != expectedTableSize {
		t.Fatalf("Table-size not the right size: (%d) != (%d)", tableSize, expectedTableSize)
	} else if len(childIfdSizes) != 0 {
		t.Fatalf("One or more child IFDs were allocated but shouldn't have been: (%d)", len(childIfdSizes))
	}

	// The ASCII value plus the rational size.
	expectedAllocatedSize := 11 + 8

	if int(allocatedDataSize) != expectedAllocatedSize {
		t.Fatalf("Allocated data size not correct: (%d)", allocatedDataSize)
	}

	expectedIfdAndDataBytes := []byte{
		// IFD table block.

		// - Tag count
		0x00, 0x04,

		// - Tags
		0x00, 0x0b, 0x00, 0x02, 0x00, 0x00, 0x00, 0x0b, 0x00, 0x00, 0x12, 0x34,
		0x00, 0xff, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x11, 0x22, 0x00, 0x00,
		0x01, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x33, 0x44, 0x55, 0x66,
		0x01, 0x3e, 0x00, 0x05, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x12, 0x3f,

		// - Next IFD offset
		0x00, 0x00, 0x00, 0x00,

		// IFD data block.

		// - The one ASCII value
		0x61, 0x73, 0x63, 0x69, 0x69, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x00,

		// - The one rational value
		0x11, 0x11, 0x22, 0x22, 0x33, 0x33, 0x44, 0x44,
	}

	if bytes.Compare(tableAndAllocated, expectedIfdAndDataBytes) != 0 {
		t.Fatalf("IFD table and allocated data not correct: %v", exifcommon.DumpBytesClauseToString(tableAndAllocated))
	}
}

func Test_IfdByteEncoder_encodeIfdToBytes_fullExif(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	ib := getExifSimpleTestIb()

	// Encode the IFD to a byte stream.

	ibe := NewIfdByteEncoder()

	// Run a simulation just to figure out the sizes.
	_, tableSize, allocatedDataSize, _, err := ibe.encodeIfdToBytes(ib, uint32(0), uint32(0), false)
	log.PanicIf(err)

	addressableOffset := ExifDefaultFirstIfdOffset + tableSize
	nextIfdOffsetToWrite := addressableOffset + allocatedDataSize

	// Run the final encode now that we can correctly assign the offsets.
	tableAndAllocated, _, _, _, err := ibe.encodeIfdToBytes(ib, addressableOffset, uint32(nextIfdOffsetToWrite), false)
	log.PanicIf(err)

	if len(tableAndAllocated) != (int(tableSize) + int(allocatedDataSize)) {
		t.Fatalf("Table-and-data size doesn't match what was expected: (%d) != (%d + %d)", len(tableAndAllocated), tableSize, allocatedDataSize)
	}

	// Wrap the IFD in a formal EXIF block.

	b := new(bytes.Buffer)

	headerBytes, err := BuildExifHeader(exifcommon.TestDefaultByteOrder, ExifDefaultFirstIfdOffset)
	log.PanicIf(err)

	_, err = b.Write(headerBytes)
	log.PanicIf(err)

	_, err = b.Write(tableAndAllocated)
	log.PanicIf(err)

	// Now, try parsing it as EXIF data, making sure to resolve (read:
	// dereference) the values (which will include the allocated ones).

	exifData := b.Bytes()
	validateExifSimpleTestIb(exifData, t)
}

func Test_IfdByteEncoder_EncodeToExifPayload(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	ib := getExifSimpleTestIb()

	// Encode the IFD to a byte stream.

	ibe := NewIfdByteEncoder()

	encodedIfds, err := ibe.EncodeToExifPayload(ib)
	log.PanicIf(err)

	// Wrap the IFD in a formal EXIF block.

	b := new(bytes.Buffer)

	headerBytes, err := BuildExifHeader(exifcommon.TestDefaultByteOrder, ExifDefaultFirstIfdOffset)
	log.PanicIf(err)

	_, err = b.Write(headerBytes)
	log.PanicIf(err)

	_, err = b.Write(encodedIfds)
	log.PanicIf(err)

	// Now, try parsing it as EXIF data, making sure to resolve (read:
	// dereference) the values (which will include the allocated ones).

	exifData := b.Bytes()
	validateExifSimpleTestIb(exifData, t)
}

func Test_IfdByteEncoder_EncodeToExif(t *testing.T) {
	ib := getExifSimpleTestIb()

	// TODO(dustin): Do a child-IFD allocation in addition to the tag allocations.

	ibe := NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	validateExifSimpleTestIb(exifData, t)
}

func Test_IfdByteEncoder_EncodeToExif_WithChildAndSibling(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintError(err)
			t.Fatalf("Test failed.")
		}
	}()

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = ib.AddStandard(0x000b, "asciivalue")
	log.PanicIf(err)

	err = ib.AddStandard(0x00ff, []uint16{0x1122})
	log.PanicIf(err)

	// Add a child IB right in the middle.

	childIb := NewIfdBuilder(im, ti, exifcommon.IfdExifStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = childIb.AddStandardWithName("ISOSpeedRatings", []uint16{0x1122})
	log.PanicIf(err)

	err = childIb.AddStandardWithName("ISOSpeed", []uint32{0x33445566})
	log.PanicIf(err)

	err = ib.AddChildIb(childIb)
	log.PanicIf(err)

	err = ib.AddStandard(0x0100, []uint32{0x33445566})
	log.PanicIf(err)

	// Add another child IB, just to ensure a little more punishment and make
	// sure we're managing our allocation offsets correctly.

	childIb2 := NewIfdBuilder(im, ti, exifcommon.IfdGpsInfoStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = childIb2.AddStandardWithName("GPSAltitudeRef", []uint8{0x11, 0x22})
	log.PanicIf(err)

	err = ib.AddChildIb(childIb2)
	log.PanicIf(err)

	err = ib.AddStandard(0x013e, []exifcommon.Rational{{Numerator: 0x11112222, Denominator: 0x33334444}})
	log.PanicIf(err)

	// Link to another IB (sibling relationship). The root/standard IFD may
	// occur twice in some JPEGs (for thumbnail or FlashPix images).

	nextIb := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = nextIb.AddStandard(0x0101, []uint32{0x11223344})
	log.PanicIf(err)

	err = nextIb.AddStandard(0x0102, []uint16{0x5566})
	log.PanicIf(err)

	ib.SetNextIb(nextIb)

	// Encode.

	ibe := NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	// Parse.

	_, index, err := Collect(im, ti, exifData)
	log.PanicIf(err)

	tagsDump := index.RootIfd.DumpTree()

	actual := strings.Join(tagsDump, "\n")

	expected :=
		`> IFD [ROOT]->[IFD]:(0) TOP
  - (0x000b)
  - (0x00ff)
  - (0x8769)
  > IFD [IFD]->[IFD/Exif]:(0) TOP
    - (0x8827)
    - (0x8833)
  < IFD [IFD]->[IFD/Exif]:(0) BOTTOM
  - (0x0100)
  - (0x8825)
  > IFD [IFD]->[IFD/GPSInfo]:(0) TOP
    - (0x0005)
  < IFD [IFD]->[IFD/GPSInfo]:(0) BOTTOM
  - (0x013e)
< IFD [ROOT]->[IFD]:(0) BOTTOM
* LINKING TO SIBLING IFD [IFD]:(1)
> IFD [ROOT]->[IFD]:(1) TOP
  - (0x0101)
  - (0x0102)
< IFD [ROOT]->[IFD]:(1) BOTTOM`

	if actual != expected {
		fmt.Printf("\n")

		fmt.Printf("Actual:\n")
		fmt.Printf("\n")
		fmt.Printf("%s\n", actual)
		fmt.Printf("\n")

		fmt.Printf("Expected:\n")
		fmt.Printf("\n")
		fmt.Printf("%s\n", expected)
		fmt.Printf("\n")

		t.Fatalf("IFD hierarchy not correct.")
	}
}

func ExampleIfdByteEncoder_EncodeToExif() {
	// Construct an IFD.

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, exifcommon.TestDefaultByteOrder)

	err = ib.AddStandardWithName("ProcessingSoftware", "asciivalue")
	log.PanicIf(err)

	err = ib.AddStandardWithName("DotRange", []uint8{0x11})
	log.PanicIf(err)

	err = ib.AddStandardWithName("SubfileType", []uint16{0x2233})
	log.PanicIf(err)

	err = ib.AddStandardWithName("ImageWidth", []uint32{0x44556677})
	log.PanicIf(err)

	err = ib.AddStandardWithName("WhitePoint", []exifcommon.Rational{{Numerator: 0x11112222, Denominator: 0x33334444}})
	log.PanicIf(err)

	err = ib.AddStandardWithName("ShutterSpeedValue", []exifcommon.SignedRational{{Numerator: 0x11112222, Denominator: 0x33334444}})
	log.PanicIf(err)

	// Encode it.

	ibe := NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	// Parse it so we can see it.

	_, index, err := Collect(im, ti, exifData)
	log.PanicIf(err)

	for i, ite := range index.RootIfd.Entries() {
		value, err := ite.Value()
		log.PanicIf(err)

		fmt.Printf("%d: %s [%v]\n", i, ite, value)
	}

	// Output:
	//
	// 0: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x000b) TAG-TYPE=[ASCII] UNIT-COUNT=(11)> [asciivalue]
	// 1: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x0150) TAG-TYPE=[BYTE] UNIT-COUNT=(1)> [[17]]
	// 2: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x00ff) TAG-TYPE=[SHORT] UNIT-COUNT=(1)> [[8755]]
	// 3: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x0100) TAG-TYPE=[LONG] UNIT-COUNT=(1)> [[1146447479]]
	// 4: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x013e) TAG-TYPE=[RATIONAL] UNIT-COUNT=(1)> [[{286335522 858997828}]]
	// 5: IfdTagEntry<TAG-IFD-PATH=[IFD] TAG-ID=(0x9201) TAG-TYPE=[SRATIONAL] UNIT-COUNT=(1)> [[{286335522 858997828}]]
}
