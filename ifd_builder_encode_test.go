package exif

import (
    "testing"
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

func Test_ByteWriter_writeAsBytes_uint8(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.writeAsBytes(uint8(0x12))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12 }) != 0 {
        t.Fatalf("uint8 not encoded correctly.")
    }
}

func Test_ByteWriter_writeAsBytes_uint16(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.writeAsBytes(uint16(0x1234))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34 }) != 0 {
        t.Fatalf("uint16 not encoded correctly.")
    }
}

func Test_ByteWriter_writeAsBytes_uint32(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.writeAsBytes(uint32(0x12345678))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34, 0x56, 0x78 }) != 0 {
        t.Fatalf("uint32 not encoded correctly.")
    }
}

func Test_ByteWriter_WriteUint16(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.WriteUint16(uint16(0x1234))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34 }) != 0 {
        t.Fatalf("uint16 not encoded correctly (as bytes).")
    }
}

func Test_ByteWriter_WriteUint32(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.WriteUint32(uint32(0x12345678))
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x12, 0x34, 0x56, 0x78 }) != 0 {
        t.Fatalf("uint32 not encoded correctly (as bytes).")
    }
}

func Test_ByteWriter_WriteFourBytes(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    err := bw.WriteFourBytes([]byte { 0x11, 0x22, 0x33, 0x44 })
    log.PanicIf(err)

    if bytes.Compare(b.Bytes(), []byte { 0x11, 0x22, 0x33, 0x44 }) != 0 {
        t.Fatalf("four-bytes not encoded correctly.")
    }
}

func Test_ByteWriter_WriteFourBytes_TooMany(t *testing.T) {
    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

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

    ib := &IfdBuilder{
        ifdName: IfdGps,
    }

    bt := &builderTag{
        tagId: 0x0000,
        value: NewIfdBuilderTagValueFromBytes([]byte { 0x12 }),
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
// TODO(dustin): !! Formally generate a BT properly and test here for every type. Make sure everything that we accomodate slices and properly encode (since things originally decode as slices)..
    childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
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

    ib := &IfdBuilder{
        ifdName: IfdGps,
    }

    bt := &builderTag{
        tagId: 0x0000,
        value: NewIfdBuilderTagValueFromBytes([]byte { 0x12, 0x34, 0x56, 0x78 }),
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
// TODO(dustin): !! Formally generate a BT properly and test here for every type. Make sure everything that we accomodate slices and properly encode (since things originally decode as slices)..
    childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
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

    ib := &IfdBuilder{
        ifdName: IfdGps,
    }

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, binary.BigEndian)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)

    bt := &builderTag{
        tagId: 0x0000,
        value: NewIfdBuilderTagValueFromBytes([]byte { 0x12, 0x34, 0x56, 0x78, 0x9A }),
    }

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
// TODO(dustin): !! Formally generate a BT properly and test here for every type. Make sure everything that we accomodate slices and properly encode (since things originally decode as slices)..
    childIfdBlock, err := ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
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

    bt = &builderTag{
        tagId: 0x0000,
        value: NewIfdBuilderTagValueFromBytes([]byte { 0xbc, 0xde, 0xf0, 0x12, 0x34 }),
    }

// TODO(dustin): !! Test with and without nextIfdOffsetToWrite.
// TODO(dustin): !! Formally generate a BT properly and test here for every type. Make sure everything that we accomodate slices and properly encode (since things originally decode as slices)..
    childIfdBlock, err = ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
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

// TODO(dustin): !! Test all types.
// TODO(dustin): !! Test specific unknown-type tags.
// TODO(dustin): !! Test what happens with unhandled unknown-type tags (though it should never get to this point in the normal workflow).
// TODO(dustin): !! Test child IFDs (may not be possible until after writing tests for higher-level IB encode).
