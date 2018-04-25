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
