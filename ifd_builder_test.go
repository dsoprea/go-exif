package exif

import (
    "testing"
    "reflect"
    "bytes"
    "path"

    "github.com/dsoprea/go-logging"
)

func TestAdd(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    originalBytes := []byte { 0x11, 0x22, 0x33 }

    bt = builderTag{
        tagId: 0x44,
        value: NewIfdBuilderTagValueFromBytes([]byte(originalBytes)),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    if ib.ii != RootIi {
        t.Fatalf("IFD name not correct.")
    } else if ib.ifdTagId != 0 {
        t.Fatalf("IFD tag-ID not correct.")
    } else if ib.byteOrder != TestDefaultByteOrder {
        t.Fatalf("IFD byte-order not correct.")
    } else if len(ib.tags) != 4 {
        t.Fatalf("IFD tag-count not correct.")
    } else if ib.existingOffset != 0 {
        t.Fatalf("IFD offset not correct.")
    } else if ib.nextIb != nil {
        t.Fatalf("Next-IFD not correct.")
    }

    tags := ib.Tags()

    if tags[0].tagId != 0x11 {
        t.Fatalf("tag (0) tag-ID not correct")
    } else if bytes.Compare(tags[0].value.Bytes(), []byte("test string")) != 0 {
        t.Fatalf("tag (0) value not correct")
    }

    if tags[1].tagId != 0x22 {
        t.Fatalf("tag (1) tag-ID not correct")
    } else if bytes.Compare(tags[1].value.Bytes(), []byte("test string2")) != 0 {
        t.Fatalf("tag (1) value not correct")
    }

    if tags[2].tagId != 0x33 {
        t.Fatalf("tag (2) tag-ID not correct")
    } else if bytes.Compare(tags[2].value.Bytes(), []byte("test string3")) != 0 {
        t.Fatalf("tag (2) value not correct")
    }

    if tags[3].tagId != 0x44 {
        t.Fatalf("tag (3) tag-ID not correct")
    } else if bytes.Compare(tags[3].value.Bytes(), originalBytes) != 0 {
        t.Fatalf("tag (3) value not correct")
    }
}

func TestSetNextIfd(t *testing.T) {
    ib1 := NewIfdBuilder(RootIi, TestDefaultByteOrder)
    ib2 := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    if ib1.nextIb != nil {
        t.Fatalf("Next-IFD for IB1 not initially terminal.")
    }

    err := ib1.SetNextIfd(ib2)
    log.PanicIf(err)

    if ib1.nextIb != ib2 {
        t.Fatalf("Next-IFD for IB1 not correct.")
    } else if ib2.nextIb != nil {
        t.Fatalf("Next-IFD for IB2 terminal.")
    }
}

func TestAddChildIb(t *testing.T) {

    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    exifIi, _ := IfdIdOrFail(IfdStandard, IfdExif)

    ibChild := NewIfdBuilder(exifIi, TestDefaultByteOrder)
    err = ib.AddChildIb(ibChild)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x11 {
        t.Fatalf("first tag not correct")
    } else if ib.tags[1].tagId != ibChild.ifdTagId {
        t.Fatalf("second tag ID does not match child-IFD tag-ID: (0x%02x) != (0x%02x)", ib.tags[1].tagId, ibChild.ifdTagId)
    } else if ib.tags[1].value.Ib() != ibChild {
        t.Fatalf("second tagvalue does not match child-IFD")
    } else if ib.tags[2].tagId != 0x22 {
        t.Fatalf("third tag not correct")
    }
}

func TestAddTagsFromExisting(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Test failure.")
        }
    }()

    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
        TagType: TypeByte,
        UnitCount: 4,
        RawValueOffset: []byte { 0x12, 0, 0, 0 },
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        ChildIfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
        TagType: TypeByte,
        UnitCount: 4,
        RawValueOffset: []byte { 0x34, 0, 0, 0 },
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, nil, nil, nil)
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
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        ChildIfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, nil, []uint16 { 0x33 }, nil)
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x33 {
        t.Fatalf("tag (1) not correct")
    } else if len(ib.tags) != 1 {
        t.Fatalf("tag count not correct")
    }
}

func TestAddTagsFromExisting__Excludes(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    entries := make([]IfdTagEntry, 3)

    entries[0] = IfdTagEntry{
        TagId: 0x11,
    }

    entries[1] = IfdTagEntry{
        TagId: 0x22,
        ChildIfdName: "some ifd",
    }

    entries[2] = IfdTagEntry{
        TagId: 0x33,
    }

    ifd := &Ifd{
        Entries: entries,
    }

    err := ib.AddTagsFromExisting(ifd, nil, nil, []uint16 { 0x11 })
    log.PanicIf(err)

    if ib.tags[0].tagId != 0x33 {
        t.Fatalf("tag not correct")
    } else if len(ib.tags) != 1 {
        t.Fatalf("tag count not correct")
    }
}

func TestFindN_First_1(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    found, err := ib.FindN(0x11, 1)
    log.PanicIf(err)

    if len(found) != 1 {
        log.Panicf("Exactly one result was not found: (%d)", len(found))
    } else if found[0] != 0 {
        log.Panicf("Result was not in the right place: (%d)", found[0])
    }

    tags := ib.Tags()
    bt = tags[found[0]]

    if bt.tagId != 0x11 {
        log.Panicf("Found entry is not correct: (0x%02x)", bt.tagId)
    }
}

func TestFindN_First_2_1Returned(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    found, err := ib.FindN(0x11, 2)
    log.PanicIf(err)

    if len(found) != 1 {
        log.Panicf("Exactly one result was not found: (%d)", len(found))
    } else if found[0] != 0 {
        log.Panicf("Result was not in the right place: (%d)", found[0])
    }

    tags := ib.Tags()
    bt = tags[found[0]]

    if bt.tagId != 0x11 {
        log.Panicf("Found entry is not correct: (0x%02x)", bt.tagId)
    }
}

func TestFindN_First_2_2Returned(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string5")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    found, err := ib.FindN(0x11, 2)
    log.PanicIf(err)

    if len(found) != 2 {
        log.Panicf("Exactly one result was not found: (%d)", len(found))
    } else if found[0] != 0 {
        log.Panicf("First result was not in the right place: (%d)", found[0])
    } else if found[1] != 3 {
        log.Panicf("Second result was not in the right place: (%d)", found[1])
    }

    tags := ib.Tags()

    bt = tags[found[0]]
    if bt.tagId != 0x11 || bytes.Compare(bt.value.Bytes(), []byte("test string")) != 0 {
        log.Panicf("Found entry 0 is not correct: (0x%02x) [%s]", bt.tagId, bt.value)
    }

    bt = tags[found[1]]
    if bt.tagId != 0x11 || bytes.Compare(bt.value.Bytes(), []byte("test string4")) != 0 {
        log.Panicf("Found entry 1 is not correct: (0x%02x) [%s]", bt.tagId, bt.value)
    }
}

func TestFindN_Middle_WithDuplicates(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string5")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string6")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    found, err := ib.FindN(0x33, 1)
    log.PanicIf(err)

    if len(found) != 1 {
        log.Panicf("Exactly one result was not found: (%d)", len(found))
    } else if found[0] != 2 {
        log.Panicf("Result was not in the right place: (%d)", found[0])
    }

    tags := ib.Tags()
    bt = tags[found[0]]

    if bt.tagId != 0x33 {
        log.Panicf("Found entry is not correct: (0x%02x)", bt.tagId)
    }
}

func TestFindN_Middle_NoDuplicates(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    found, err := ib.FindN(0x33, 1)
    log.PanicIf(err)

    if len(found) != 1 {
        log.Panicf("Exactly one result was not found: (%d)", len(found))
    } else if found[0] != 2 {
        log.Panicf("Result was not in the right place: (%d)", found[0])
    }

    tags := ib.Tags()
    bt = tags[found[0]]

    if bt.tagId != 0x33 {
        log.Panicf("Found entry is not correct: (0x%02x)", bt.tagId)
    }
}

func TestFindN_Miss(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    found, err := ib.FindN(0x11, 1)
    log.PanicIf(err)

    if len(found) != 0 {
        t.Fatalf("Expected empty results.")
    }
}

func TestFind_Hit(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    position, err := ib.Find(0x33)
    log.PanicIf(err)

    if position != 2 {
        log.Panicf("Result was not in the right place: (%d)", position)
    }

    tags := ib.Tags()
    bt = tags[position]

    if bt.tagId != 0x33 {
        log.Panicf("Found entry is not correct: (0x%02x)", bt.tagId)
    }
}

func TestFind_Miss(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    _, err = ib.Find(0x99)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestReplace(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    currentIds := make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-replace tags are not correct.")
    }

    bt = builderTag{
        tagId: 0x99,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Replace(0x22, bt)
    log.PanicIf(err)

    currentIds = make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x99, 0x33 }, currentIds) == false {
        t.Fatalf("Post-replace tags are not correct.")
    }
}

func TestReplaceN(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    currentIds := make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-replace tags are not correct.")
    }

    bt = builderTag{
        tagId: 0xA9,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.ReplaceAt(1, bt)
    log.PanicIf(err)

    currentIds = make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0xA9, 0x33 }, currentIds) == false {
        t.Fatalf("Post-replace tags are not correct.")
    }
}

func TestDeleteFirst(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)


    if len(ib.Tags()) != 4 {
        t.Fatalf("Pre-delete tag count not correct.")
    }

    currentIds := make([]uint16, 4)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-delete tags not correct.")
    }


    err = ib.DeleteFirst(0x22)
    log.PanicIf(err)

    if len(ib.Tags()) != 3 {
        t.Fatalf("Post-delete (1) tag count not correct.")
    }

    currentIds = make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete (1) tags not correct.")
    }


    err = ib.DeleteFirst(0x22)
    log.PanicIf(err)

    if len(ib.Tags()) != 2 {
        t.Fatalf("Post-delete (2) tag count not correct.")
    }

    currentIds = make([]uint16, 2)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete (2) tags not correct.")
    }


    err = ib.DeleteFirst(0x22)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestDeleteN(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)


    if len(ib.Tags()) != 4 {
        t.Fatalf("Pre-delete tag count not correct.")
    }

    currentIds := make([]uint16, 4)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-delete tags not correct.")
    }


    err = ib.DeleteN(0x22, 1)
    log.PanicIf(err)

    if len(ib.Tags()) != 3 {
        t.Fatalf("Post-delete (1) tag count not correct.")
    }

    currentIds = make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete (1) tags not correct.")
    }


    err = ib.DeleteN(0x22, 1)
    log.PanicIf(err)

    if len(ib.Tags()) != 2 {
        t.Fatalf("Post-delete (2) tag count not correct.")
    }

    currentIds = make([]uint16, 2)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete (2) tags not correct.")
    }


    err = ib.DeleteN(0x22, 1)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestDeleteN_Two(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)


    if len(ib.Tags()) != 4 {
        t.Fatalf("Pre-delete tag count not correct.")
    }

    currentIds := make([]uint16, 4)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-delete tags not correct.")
    }


    err = ib.DeleteN(0x22, 2)
    log.PanicIf(err)

    if len(ib.Tags()) != 2 {
        t.Fatalf("Post-delete tag count not correct.")
    }

    currentIds = make([]uint16, 2)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete tags not correct.")
    }


    err = ib.DeleteFirst(0x22)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestDeleteAll(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    bt := builderTag{
        tagId: 0x11,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string")),
    }

    err := ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string2")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x22,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string3")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)

    bt = builderTag{
        tagId: 0x33,
        value: NewIfdBuilderTagValueFromBytes([]byte("test string4")),
    }

    err = ib.Add(bt)
    log.PanicIf(err)


    if len(ib.Tags()) != 4 {
        t.Fatalf("Pre-delete tag count not correct.")
    }

    currentIds := make([]uint16, 4)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-delete tags not correct.")
    }


    n, err := ib.DeleteAll(0x22)
    log.PanicIf(err)

    if n != 2 {
        t.Fatalf("Returned delete tag count not correct.")
    } else if len(ib.Tags()) != 2 {
        t.Fatalf("Post-delete tag count not correct.")
    }

    currentIds = make([]uint16, 2)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x33 }, currentIds) == false {
        t.Fatalf("Post-delete tags not correct.")
    }


    err = ib.DeleteFirst(0x22)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestNewIfdBuilderFromExistingChain(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Test failure.")
        }
    }()

    e := NewExif()

    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    exifData, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    _, index, err := e.Collect(exifData)
    log.PanicIf(err)

    ib := NewIfdBuilderFromExistingChain(index.RootIfd, exifData)
    lines := ib.DumpToStrings()

    expected := []string {
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(0) TAG=[0x10f]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(1) TAG=[0x110]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(2) TAG=[0x112]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(3) TAG=[0x11a]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(4) TAG=[0x11b]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(5) TAG=[0x128]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(6) TAG=[0x132]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(7) TAG=[0x13b]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(8) TAG=[0x213]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[] INDEX=(9) TAG=[0x8298]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[Exif] INDEX=(10) TAG=[0x8769]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(0) TAG=[0x829a]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(1) TAG=[0x829d]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(2) TAG=[0x8822]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(3) TAG=[0x8827]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(4) TAG=[0x8830]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(5) TAG=[0x8832]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(6) TAG=[0x9000]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(7) TAG=[0x9003]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(8) TAG=[0x9004]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(9) TAG=[0x9101]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(10) TAG=[0x9201]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(11) TAG=[0x9202]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(12) TAG=[0x9204]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(13) TAG=[0x9207]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(14) TAG=[0x9209]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(15) TAG=[0x920a]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(16) TAG=[0x927c]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(17) TAG=[0x9286]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(18) TAG=[0x9290]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(19) TAG=[0x9291]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(20) TAG=[0x9292]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(21) TAG=[0xa000]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(22) TAG=[0xa001]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(23) TAG=[0xa002]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(24) TAG=[0xa003]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(25) TAG=[0xa20e]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(26) TAG=[0xa20f]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(27) TAG=[0xa210]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(28) TAG=[0xa401]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(29) TAG=[0xa402]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(30) TAG=[0xa403]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(31) TAG=[0xa406]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(32) TAG=[0xa430]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(33) TAG=[0xa431]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(34) TAG=[0xa432]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(35) TAG=[0xa434]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(36) TAG=[0xa435]",
        "<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[Iop] INDEX=(37) TAG=[0xa005]",
        "<PARENTS=[IFD->Exif] IFD-NAME=[Iop]> IFD-TAG-ID=(0xa005) CHILD-IFD=[] INDEX=(0) TAG=[0x01]",
        "<PARENTS=[IFD->Exif] IFD-NAME=[Iop]> IFD-TAG-ID=(0xa005) CHILD-IFD=[] INDEX=(1) TAG=[0x02]",
        "<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x00) CHILD-IFD=[GPSInfo] INDEX=(11) TAG=[0x8825]",
        "<PARENTS=[IFD] IFD-NAME=[GPSInfo]> IFD-TAG-ID=(0x8825) CHILD-IFD=[] INDEX=(0) TAG=[0x00]",
    }

    if reflect.DeepEqual(lines, expected) == false {
        t.Fatalf("IB did not [correctly] duplicate the IFD structure")
    }
}

// TODO(dustin): !! Test with an actual GPS-attached image.


func TestNewIfdBuilderWithExistingIfd(t *testing.T) {
    ii, _ := IfdIdOrFail(IfdStandard, IfdGps)
    tagId := IfdTagIdWithIdentityOrFail(ii)

    parentIfd := &Ifd{
        Name: IfdStandard,
    }

    ifd := &Ifd{
        Name: IfdGps,
        ByteOrder: TestDefaultByteOrder,
        Offset: 0x123,
        ParentIfd: parentIfd,
    }

    ib := NewIfdBuilderWithExistingIfd(ifd)

    if ib.ii.IfdName != ifd.Name {
        t.Fatalf("IFD-name not correct.")
    } else if ib.ifdTagId != tagId {
        t.Fatalf("IFD tag-ID not correct.")
    } else if ib.byteOrder != ifd.ByteOrder {
        t.Fatalf("IFD byte-order not correct.")
    } else if ib.existingOffset != ifd.Offset {
        t.Fatalf("IFD offset not correct.")
    }
}

func TestNewStandardBuilderTagFromConfig_OneUnit(t *testing.T) {
    bt := NewStandardBuilderTagFromConfig(ExifIi, uint16(0x8833), TestDefaultByteOrder, []uint32 { uint32(0x1234) })

    if bt.ii != ExifIi {
        t.Fatalf("II in builderTag not correct")
    } else if bt.tagId != 0x8833 {
        t.Fatalf("tag-ID not correct")
    } else if bytes.Compare(bt.value.Bytes(), []byte { 0x0, 0x0, 0x12, 0x34, }) != 0 {
        t.Fatalf("value not correct")
    }
}

func TestNewStandardBuilderTagFromConfig_TwoUnits(t *testing.T) {
    bt := NewStandardBuilderTagFromConfig(ExifIi, uint16(0x8833), TestDefaultByteOrder, []uint32 { uint32(0x1234), uint32(0x5678) })

    if bt.ii != ExifIi {
        t.Fatalf("II in builderTag not correct")
    } else if bt.tagId != 0x8833 {
        t.Fatalf("tag-ID not correct")
    } else if bytes.Compare(bt.value.Bytes(), []byte {
            0x0, 0x0, 0x12, 0x34,
            0x0, 0x0, 0x56, 0x78, }) != 0 {
        t.Fatalf("value not correct")
    }
}

func TestNewStandardBuilderTagFromConfigWithName(t *testing.T) {
    bt := NewStandardBuilderTagFromConfigWithName(ExifIi, "ISOSpeed", TestDefaultByteOrder, []uint32 { uint32(0x1234), uint32(0x5678) })

    if bt.ii != ExifIi {
        t.Fatalf("II in builderTag not correct")
    } else if bt.tagId != 0x8833 {
        t.Fatalf("tag-ID not correct")
    } else if bytes.Compare(bt.value.Bytes(), []byte {
            0x0, 0x0, 0x12, 0x34,
            0x0, 0x0, 0x56, 0x78, }) != 0 {
        t.Fatalf("value not correct")
    }
}

func TestAddFromConfigWithName(t *testing.T) {
    ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

    err := ib.AddFromConfigWithName("ProcessingSoftware", "some software")
    log.PanicIf(err)

    if len(ib.tags) != 1 {
        t.Fatalf("Exactly one tag was not found: (%d)", len(ib.tags))
    }

    bt := ib.tags[0]

    if bt.ii != RootIi {
        t.Fatalf("II not correct: %s", bt.ii)
    } else if bt.tagId != 0x000b {
        t.Fatalf("Tag-ID not correct: (0x%02x)", bt.tagId)
    }

    s := string(bt.value.Bytes())

    if s != "some software\000" {
        t.Fatalf("Value not correct: (%d) [%s]", len(s), s)
    }
}
