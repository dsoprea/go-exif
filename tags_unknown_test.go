package exif

import (
    "testing"
    "bytes"

    "github.com/dsoprea/go-logging"
)

func TestUndefinedValue_ExifVersion(t *testing.T) {
    byteOrder := TestDefaultByteOrder
    ii := ExifIi


    // Create our unknown-type tag's value using the fact that we know it's a
    // non-null-terminated string.

    ve := NewValueEncoder(byteOrder)

    tt := NewTagType(TypeAsciiNoNul, byteOrder)
    valueString := "0230"

    ed, err := ve.EncodeWithType(tt, valueString)
    log.PanicIf(err)


    // Create the tag using the official "unknown" type now that we already
    // have the bytes.

    encodedValue := NewIfdBuilderTagValueFromBytes(ed.Encoded)

    bt := &BuilderTag{
        ii: ii,
        tagId: 0x9000,
        typeId: TypeUndefined,
        value: encodedValue,
    }


    // Stage the build.

    ibe := NewIfdByteEncoder()
    ib := NewIfdBuilder(ii, byteOrder)

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, byteOrder)

    addressableOffset := uint32(0x1234)
    ida := newIfdDataAllocator(addressableOffset)


    // Encode.

    _, err = ibe.encodeTagToBytes(ib, bt, bw, ida, uint32(0))
    log.PanicIf(err)

    tagBytes := b.Bytes()

    if len(tagBytes) != 12 {
        t.Fatalf("Tag not encoded to the right number of bytes: (%d)", len(tagBytes))
    }

    ite, err := ParseOneTag(ii, byteOrder, tagBytes)
    log.PanicIf(err)

    if ite.TagId != 0x9000 {
        t.Fatalf("Tag-ID not correct: (0x%02x)", ite.TagId)
    } else if ite.TagIndex != 0 {
        t.Fatalf("Tag index not correct: (%d)", ite.TagIndex)
    } else if ite.TagType != TypeUndefined {
        t.Fatalf("Tag type not correct: (%d)", ite.TagType)
    } else if ite.UnitCount != (uint32(len(valueString))) {
        t.Fatalf("Tag unit-count not correct: (%d)", ite.UnitCount)
    } else if bytes.Compare(ite.RawValueOffset, []byte { '0', '2', '3', '0' }) != 0 {
        t.Fatalf("Tag's value (as raw bytes) is not correct: [%x]", ite.RawValueOffset)
    } else if ite.ChildIfdName != "" {
        t.Fatalf("Tag's IFD-name should be empty: [%s]", ite.ChildIfdName)
    } else if ite.Ii != ii {
        t.Fatalf("Tag's parent IFD is not correct: %v", ite.Ii)
    }
}

// TODO(dustin): !! Add tests for remaining, well-defined unknown
// TODO(dustin): !! Test what happens with unhandled unknown-type tags (though it should never get to this point in the normal workflow).
