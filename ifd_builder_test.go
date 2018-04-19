package exif

import (
    "testing"
    "reflect"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)


// TODO(dustin): !! We need to have knowledge of the types so that we can validate or induce extra info for the adds.


func TestAdd(t *testing.T) {
    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    bt := builderTag{
        tagId: 0x11,
        value: "test string",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x22,
        value: "test string2",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string3",
    }

    ib.Add(bt)

    originalShorts := []uint16 { 0x111, 0x222, 0x333 }

    bt = builderTag{
        tagId: 0x44,
        value: originalShorts,
    }

    ib.Add(bt)

    if ib.ifdName != IfdStandard {
        t.Fatalf("IFD name not correct.")
    } else if ib.ifdTagId != 0 {
        t.Fatalf("IFD tag-ID not correct.")
    } else if ib.byteOrder != binary.BigEndian {
        t.Fatalf("IFD byte-order not correct.")
    } else if len(ib.tags) != 4 {
        t.Fatalf("IFD tag-count not correct.")
    } else if ib.existingOffset != 0 {
        t.Fatalf("IFD offset not correct.")
    } else if ib.nextIfd != nil {
        t.Fatalf("Next-IFD not correct.")
    }

    tags := ib.Tags()

    if tags[0].tagId != 0x11 {
        t.Fatalf("tag (0) tag-ID not correct")
    } else if tags[0].value != "test string" {
        t.Fatalf("tag (0) value not correct")
    }

    if tags[1].tagId != 0x22 {
        t.Fatalf("tag (1) tag-ID not correct")
    } else if tags[1].value != "test string2" {
        t.Fatalf("tag (1) value not correct")
    }

    if tags[2].tagId != 0x33 {
        t.Fatalf("tag (2) tag-ID not correct")
    } else if tags[2].value != "test string3" {
        t.Fatalf("tag (2) value not correct")
    }

    if tags[3].tagId != 0x44 {
        t.Fatalf("tag (3) tag-ID not correct")
    } else if reflect.DeepEqual(tags[3].value.([]uint16), originalShorts) != true {
        t.Fatalf("tag (3) value not correct")
    }
}

func TestSetNextIfd(t *testing.T) {
    ib1 := NewIfdBuilder(IfdStandard, binary.BigEndian)
    ib2 := NewIfdBuilder(IfdStandard, binary.BigEndian)

    if ib1.nextIfd != nil {
        t.Fatalf("Next-IFD for IB1 not initially terminal.")
    }

    err := ib1.SetNextIfd(ib2)
    log.PanicIf(err)

    if ib1.nextIfd != ib2 {
        t.Fatalf("Next-IFD for IB1 not correct.")
    } else if ib2.nextIfd != nil {
        t.Fatalf("Next-IFD for IB2 terminal.")
    }
}

func TestAddChildIfd(t *testing.T) {

    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    bt := builderTag{
        tagId: 0x11,
        value: "test string",
    }

    ib.Add(bt)

    ibChild := NewIfdBuilder(IfdExif, binary.BigEndian)
    err := ib.AddChildIfd(ibChild)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: "test string",
    }

    ib.Add(bt)

    if ib.tags[0].tagId != 0x11 {
        t.Fatalf("first tag not correct")
    } else if ib.tags[1].tagId != ibChild.ifdTagId {
        t.Fatalf("second tag ID does not match child-IFD tag-ID")
    } else if ib.tags[1].value != ibChild {
        t.Fatalf("second tagvalue does not match child-IFD")
    } else if ib.tags[2].tagId != 0x22 {
        t.Fatalf("third tag not correct")
    }
}

func TestAddTagsFromExisting(t *testing.T) {
    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        IfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, nil, nil)
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x11 {
        t.Fatalf("tag (0) not correct")
    } else if ib.tags[1].tagId != 0x33 {
        t.Fatalf("tag (1) not correct")
    } else if len(ib.tags) != 2 {
        t.Fatalf("tag count not correct")
    }
}

func TestAddTagsFromExisting__Includes(t *testing.T) {
    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        IfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, []uint16 { 0x33 }, nil)
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x33 {
        t.Fatalf("tag (1) not correct")
    } else if len(ib.tags) != 1 {
        t.Fatalf("tag count not correct")
    }
}

func TestAddTagsFromExisting__Excludes(t *testing.T) {
    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        IfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, nil, []uint16 { 0x11 })
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x33 {
        t.Fatalf("tag not correct")
    } else if len(ib.tags) != 1 {
        t.Fatalf("tag count not correct")
    }
}
