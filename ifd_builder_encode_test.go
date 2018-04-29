package exif

import (
    "testing"
    "bytes"
    // "fmt"

    "github.com/dsoprea/go-logging"
)

func Test_ByteWriter_writeAsBytes_uint8(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.writeAsBytes(uint8(0x12))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12 }) != 0 {
        t.Fatalf("uint8 not encoded correctly.")
    }
}

func Test_ByteWriter_writeAsBytes_uint16(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.writeAsBytes(uint16(0x1234))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34 }) != 0 {
        t.Fatalf("uint16 not encoded correctly.")
    }
}

func Test_ByteWriter_writeAsBytes_uint32(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.writeAsBytes(uint32(0x12345678))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34, 0x56, 0x78 }) != 0 {
        t.Fatalf("uint32 not encoded correctly.")
    }
}

func Test_ByteWriter_WriteUint16(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.WriteUint16(uint16(0x1234))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34 }) != 0 {
        t.Fatalf("uint16 not encoded correctly (as bytes).")
    }
}

func Test_ByteWriter_WriteUint32(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.WriteUint32(uint32(0x12345678))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34, 0x56, 0x78 }) != 0 {
        t.Fatalf("uint32 not encoded correctly (as bytes).")
    }
}

func Test_ByteWriter_WriteFourBytes(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.WriteFourBytes([]byte { 0x11, 0x22, 0x33, 0x44 })
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x11, 0x22, 0x33, 0x44 }) != 0 {
        t.Fatalf("four-bytes not encoded correctly.")
    }
}

func Test_ByteWriter_WriteFourBytes_TooMany(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    err := bw.WriteFourBytes([]byte { 0x11, 0x22, 0x33, 0x44, 0x55 })
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

    data := []byte { 0x1, 0x2, 0x3 }
    offset, err := ida.Allocate(data)
    log.PanicIf(err)

    expected := uint32(addressableOffset + 0)
    if offset != expected {
        t.Fatalf("offset not bumped correctly (2): (%d) != (%d)", offset, expected)
    } else if ida.NextOffset() != offset + uint32(3) {
        t.Fatalf("position counter not advanced properly")
    } else if bytes.Compare(ida.Bytes(), []byte { 0x1, 0x2, 0x3 }) != 0 {
        t.Fatalf("buffer not correct after write (1)")
    }

    data = []byte { 0x4, 0x5, 0x6 }
    offset, err = ida.Allocate(data)
    log.PanicIf(err)

    expected = uint32(addressableOffset + 3)
    if offset != expected {
        t.Fatalf("offset not bumped correctly (3): (%d) != (%d)", offset, expected)
    } else if ida.NextOffset() != offset + uint32(3) {
        t.Fatalf("position counter not advanced properly")
    } else if bytes.Compare(ida.Bytes(), []byte { 0x1, 0x2, 0x3, 0x4, 0x5, 0x6 }) != 0 {
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

    data := []byte { 0x1, 0x2, 0x3 }
    offset, err := ida.Allocate(data)
    log.PanicIf(err)

    expected := uint32(addressableOffset + 0)
    if offset != expected {
        t.Fatalf("offset not bumped correctly (2): (%d) != (%d)", offset, expected)
    } else if ida.NextOffset() != offset + uint32(3) {
        t.Fatalf("position counter not advanced properly")
    } else if bytes.Compare(ida.Bytes(), []byte { 0x1, 0x2, 0x3 }) != 0 {
        t.Fatalf("buffer not correct after write (1)")
    }

    data = []byte { 0x4, 0x5, 0x6 }
    offset, err = ida.Allocate(data)
    log.PanicIf(err)

    expected = uint32(addressableOffset + 3)
    if offset != expected {
        t.Fatalf("offset not bumped correctly (3): (%d) != (%d)", offset, expected)
    } else if ida.NextOffset() != offset + uint32(3) {
        t.Fatalf("position counter not advanced properly")
    } else if bytes.Compare(ida.Bytes(), []byte { 0x1, 0x2, 0x3, 0x4, 0x5, 0x6 }) != 0 {
        t.Fatalf("buffer not correct after write (2)")
    }
}


func Test_IfdByteEncoder__Arithmetic(t *testing.T) {
    ibe := NewIfdByteEncoder()

    if (ibe.TableSize(1) - ibe.TableSize(0)) != ibe.EntrySize() {
        t.Fatalf("table-size/entry-size not consistent (1)")
    } else if (ibe.TableSize(11) - ibe.TableSize(10)) != ibe.EntrySize() {
        t.Fatalf("table-size/entry-size not consistent (2)")
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_embedded1(t *testing.T) {
    ibe := NewIfdByteEncoder()

    gpsIi, _ := IfdIdOrFail(IfdStandard, IfdGps)

    ib := &IfdBuilder{
        ii: gpsIi,
    }

    bt := NewBuilderTagFromConfig(GpsIi, 0x0000, TestDefaultByteOrder, []uint8 { uint8(0x12) })

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, uint32(0))
    log.PanicIf(err)

    if childIfdBlock != nil {
        t.Fatalf("no child-IFDs were expected to be allocated")
    } else if bytes.Compare(b.Bytes(), []byte { 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x12, 0x00, 0x00, 0x00 }) != 0 {
        t.Fatalf("encoded tag-entry bytes not correct")
    } else if ida.NextOffset() != addressableOffset {
        t.Fatalf("allocation was done but not expected")
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_embedded2(t *testing.T) {
    ibe := NewIfdByteEncoder()

    gpsIi, _ := IfdIdOrFail(IfdStandard, IfdGps)

    ib := &IfdBuilder{
        ii: gpsIi,
    }

    bt := NewBuilderTagFromConfig(GpsIi, 0x0000, TestDefaultByteOrder, []uint8 { uint8(0x12), uint8(0x34), uint8(0x56), uint8(0x78) })

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, uint32(0))
    log.PanicIf(err)

    if childIfdBlock != nil {
        t.Fatalf("no child-IFDs were expected to be allocated")
    } else if bytes.Compare(b.Bytes(), []byte { 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x12, 0x34, 0x56, 0x78 }) != 0 {
        t.Fatalf("encoded tag-entry bytes not correct")
    } else if ida.NextOffset() != addressableOffset {
        t.Fatalf("allocation was done but not expected")
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_bytes_allocated(t *testing.T) {
    ibe := NewIfdByteEncoder()

    gpsIi, _ := IfdIdOrFail(IfdStandard, IfdGps)

    ib := &IfdBuilder{
        ii: gpsIi,
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

    bt := NewBuilderTagFromConfig(GpsIi, 0x0000, TestDefaultByteOrder, []uint8 { uint8(0x12), uint8(0x34), uint8(0x56), uint8(0x78), uint8(0x9a) })

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, uint32(0))
    log.PanicIf(err)

    if childIfdBlock != nil {
        t.Fatalf("no child-IFDs were expected to be allocated (1)")
    } else if bytes.Compare(b.Bytes(), []byte { 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x34 }) != 0 {
        t.Fatalf("encoded tag-entry bytes not correct (1)")
    } else if ida.NextOffset() != addressableOffset + uint32(5) {
        t.Fatalf("allocation offset not expected (1)")
    } else if bytes.Compare(ida.Bytes(), []byte { 0x12, 0x34, 0x56, 0x78, 0x9A }) != 0 {
        t.Fatalf("allocated data not correct (1)")
    }

    // Test that another allocation encodes to the new offset.

    bt = NewBuilderTagFromConfig(GpsIi, 0x0000, TestDefaultByteOrder, []uint8 { uint8(0xbc), uint8(0xde), uint8(0xf0), uint8(0x12), uint8(0x34) })

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
    childIfdBlock, err = ibe.encodeTagToBytes(ib, &bt, bw, ida, uint32(0))
    log.PanicIf(err)

    if childIfdBlock != nil {
        t.Fatalf("no child-IFDs were expected to be allocated (2)")
    } else if bytes.Compare(b.Bytes(), []byte {
                0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x34, // Tag 1
                0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x12, 0x39, // Tag 2
            }) != 0 {
        t.Fatalf("encoded tag-entry bytes not correct (2)")
    } else if ida.NextOffset() != addressableOffset + uint32(10) {
        t.Fatalf("allocation offset not expected (2)")
    } else if bytes.Compare(ida.Bytes(), []byte {
                0x12, 0x34, 0x56, 0x78, 0x9A,
                0xbc, 0xde, 0xf0, 0x12, 0x34,
            }) != 0 {
        t.Fatalf("allocated data not correct (2)")
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_childIfd__withoutAllocate(t *testing.T) {
    ibe := NewIfdByteEncoder()

    ib := &IfdBuilder{
        ii: RootIi,
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

    childIb := NewIfdBuilder(ExifIi, TestDefaultByteOrder)
    bt := NewBuilderTagFromConfig(RootIi, IfdExifId, TestDefaultByteOrder, childIb)

    nextIfdOffsetToWrite := uint32(0)
    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, nextIfdOffsetToWrite)
    log.PanicIf(err)

    if childIfdBlock != nil {
        t.Fatalf("no child-IFDs were expected to be allocated")
    } else if bytes.Compare(b.Bytes(), []byte { 0x87, 0x69, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00 }) != 0 {
        t.Fatalf("encoded tag-entry with child-IFD not correct")
    } else if ida.NextOffset() != addressableOffset {
        t.Fatalf("allocation offset not expected")
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_childIfd__withAllocate(t *testing.T) {
    // Create a child IFD (represented by an IB instance) that we can allocate
    // space for and then attach to a tag (which would normally be an entry,
    // then, in a higher IFD).

    childIb := NewIfdBuilder(ExifIi, TestDefaultByteOrder)

    childIbTestTag := builderTag{
        ii: ExifIi,
        tagId: 0x8822,
        value: NewIfdBuilderTagValueFromBytes([]byte { 0x12, 0x34 }),
    }

    childIb.Add(childIbTestTag)

    // Formally compose the tag that refers to it.

    bt := NewBuilderTagFromConfig(RootIi, IfdExifId, TestDefaultByteOrder, childIb)

    // Encode the tag. Since we've actually provided an offset at which we can
    // allocate data, the child-IFD will automatically be encoded, allocated,
    // and installed into the allocated-data block (which will follow the IFD
    // block/table in the file).

    ibe := NewIfdByteEncoder()

    ib := &IfdBuilder{
        ii: RootIi,
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

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

    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, nextIfdOffsetToWrite)
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

    iteV, err := ParseOneTag(RootIi, TestDefaultByteOrder, tagBytes)
    log.PanicIf(err)

    if iteV.TagId != IfdExifId {
        t.Fatalf("IFD first tag-ID not correct: (0x%02x)", iteV.TagId)
    } else if iteV.TagIndex != 0 {
        t.Fatalf("IFD first tag index not correct: (%d)", iteV.TagIndex)
    } else if iteV.TagType != TypeLong {
        t.Fatalf("IFD first tag type not correct: (%d)", iteV.TagType)
    } else if iteV.UnitCount != 1 {
        t.Fatalf("IFD first tag unit-count not correct: (%d)", iteV.UnitCount)
    } else if iteV.ValueOffset != nextIfdOffsetToWrite {
        t.Fatalf("IFD's child-IFD offset (as offset) is not correct: (%d) != (%d)", iteV.ValueOffset, nextIfdOffsetToWrite)
    } else if bytes.Compare(iteV.RawValueOffset, []byte { 0x0, 0x0, 0x07, 0xd0 }) != 0 {
        t.Fatalf("IFD's child-IFD offset (as raw bytes) is not correct: [%x]", iteV.RawValueOffset)
    } else if iteV.ChildIfdName != IfdExif {
        t.Fatalf("IFD first tag IFD-name name not correct: [%s]", iteV.ChildIfdName)
    } else if iteV.Ii != RootIi {
        t.Fatalf("IFD first tag parent IFD not correct: %v", iteV.Ii)
    }


// TODO(dustin): Test writing some tags that require allocation.
// TODO(dustin): Do an child-IFD allocation in addition to some tag allocations, and vice-verse.


    // Validate the child's raw IFD bytes.

    childNextIfdOffset, childEntries, err := ParseOneIfd(ExifIi, TestDefaultByteOrder, childIfdBlock, nil)
    log.PanicIf(err)

    if childNextIfdOffset != uint32(0) {
        t.Fatalf("Child IFD: Next IFD offset should be (0): (0x%08x)", childNextIfdOffset)
    } else if len(childEntries) != 1 {
        t.Fatalf("Child IFD: Expected exactly one entry: (%d)", len(childEntries))
    }

    ite := childEntries[0]

    if ite.TagId != 0x8822 {
        t.Fatalf("Child IFD first tag-ID not correct: (0x%02x)", ite.TagId)
    } else if ite.TagIndex != 0 {
        t.Fatalf("Child IFD first tag index not correct: (%d)", ite.TagIndex)
    } else if ite.TagType != TypeShort {
        t.Fatalf("Child IFD first tag type not correct: (%d)", ite.TagType)
    } else if ite.UnitCount != 1 {
        t.Fatalf("Child IFD first tag unit-count not correct: (%d)", ite.UnitCount)
    } else if ite.ValueOffset != 0x12340000 {
        t.Fatalf("Child IFD first tag value value (as offset) not correct: (0x%02x)", ite.ValueOffset)
    } else if bytes.Compare(ite.RawValueOffset, []byte { 0x12, 0x34, 0x0, 0x0 }) != 0 {
        t.Fatalf("Child IFD first tag value value (as raw bytes) not correct: [%v]", ite.RawValueOffset)
    } else if ite.ChildIfdName != "" {
        t.Fatalf("Child IFD first tag IFD-name name not empty: [%s]", ite.ChildIfdName)
    } else if ite.Ii != ExifIi {
        t.Fatalf("Child IFD first tag parent IFD not correct: %v", ite.Ii)
    }
}

func Test_IfdByteEncoder_encodeTagToBytes_allocate(t *testing.T) {
    // Encode the tag. Since we've actually provided an offset at which we can
    // allocate data, the child-IFD will automatically be encoded, allocated,
    // and installed into the allocated-data block (which will follow the IFD
    // block/table in the file).

    ibe := NewIfdByteEncoder()

    ib := &IfdBuilder{
        ii: RootIi,
    }

    valueString := "testvalue"
    bt := NewBuilderTagFromConfig(RootIi, 0x000b, TestDefaultByteOrder, valueString)

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, TestDefaultByteOrder)

    // addressableOffset is the offset of where large data can be allocated
    // (which follows the IFD table/block). Large, in that it can't be stored
    // in the table itself. Just used for arithmetic. This is just where the
    // data for the current IFD can be written. It's not absolute for the EXIF
    // data in general.
    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

    childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, uint32(0))
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

    ite, err := ParseOneTag(RootIi, TestDefaultByteOrder, tagBytes)
    log.PanicIf(err)

// TODO(dustin): !! When we eventually start allocating values and child-IFDs, be careful that the offsets are calculated correctly.

    if ite.TagId != 0x000b {
        t.Fatalf("Tag-ID not correct: (0x%02x)", ite.TagId)
    } else if ite.TagIndex != 0 {
        t.Fatalf("Tag index not correct: (%d)", ite.TagIndex)
    } else if ite.TagType != TypeAscii {
        t.Fatalf("Tag type not correct: (%d)", ite.TagType)
    } else if ite.UnitCount != (uint32(len(valueString) + 1)) {
        t.Fatalf("Tag unit-count not correct: (%d)", ite.UnitCount)
    } else if ite.ValueOffset != addressableOffset {
        t.Fatalf("Tag's value (as offset) is not correct: (%d) != (%d)", ite.ValueOffset, addressableOffset)
    } else if bytes.Compare(ite.RawValueOffset, []byte { 0x0, 0x0, 0x12, 0x34 }) != 0 {
        t.Fatalf("Tag's value (as raw bytes) is not correct: [%x]", ite.RawValueOffset)
    } else if ite.ChildIfdName != "" {
        t.Fatalf("Tag's IFD-name should be empty: [%s]", ite.ChildIfdName)
    } else if ite.Ii != RootIi {
        t.Fatalf("Tag's parent IFD is not correct: %v", ite.Ii)
    }

    expectedBuffer := bytes.NewBufferString(valueString)
    expectedBuffer.Write([]byte { 0x0 })
    expectedBytes := expectedBuffer.Bytes()

    allocatedBytes := ida.Bytes()

    if bytes.Compare(allocatedBytes, expectedBytes) != 0 {
        t.Fatalf("Allocated bytes not correct: %v != %v", allocatedBytes, expectedBytes)
    }
}

// TODO(dustin): !! Test all types.
// TODO(dustin): !! Test specific unknown-type tags.
// TODO(dustin): !! Test what happens with unhandled unknown-type tags (though it should never get to this point in the normal workflow).

// TODO(dustin): !! Once we can correctly build a complete EXIF, test by trying to parse with an Exif instance.
