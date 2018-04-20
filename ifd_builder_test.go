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

func TestFindN_First_1(t *testing.T) {
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

    bt = builderTag{
        tagId: 0x11,
        value: "test string4",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x11,
        value: "test string5",
    }

    ib.Add(bt)

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
    if bt.tagId != 0x11 || bt.value != "test string" {
        log.Panicf("Found entry 0 is not correct: (0x%02x) [%s]", bt.tagId, bt.value)
    }

    bt = tags[found[1]]
    if bt.tagId != 0x11 || bt.value != "test string4" {
        log.Panicf("Found entry 1 is not correct: (0x%02x) [%s]", bt.tagId, bt.value)
    }
}

func TestFindN_Middle_WithDuplicates(t *testing.T) {
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

    bt = builderTag{
        tagId: 0x11,
        value: "test string4",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x11,
        value: "test string5",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string6",
    }

    ib.Add(bt)

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

    bt = builderTag{
        tagId: 0x11,
        value: "test string4",
    }

    ib.Add(bt)

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
    ib := NewIfdBuilder(IfdStandard, binary.BigEndian)

    found, err := ib.FindN(0x11, 1)
    log.PanicIf(err)

    if len(found) != 0 {
        t.Fatalf("Expected empty results.")
    }
}

func TestFind_Hit(t *testing.T) {
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

    bt = builderTag{
        tagId: 0x11,
        value: "test string4",
    }

    ib.Add(bt)

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

    bt = builderTag{
        tagId: 0x11,
        value: "test string4",
    }

    ib.Add(bt)

    _, err := ib.Find(0x99)
    if err == nil {
        t.Fatalf("Expected an error.")
    } else if log.Is(err, ErrTagEntryNotFound) == false {
        log.Panic(err)
    }
}

func TestReplace(t *testing.T) {
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

    currentIds := make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-replace tags are not correct.")
    }

    bt = builderTag{
        tagId: 0x99,
        value: "test string4",
    }

    err := ib.Replace(0x22, bt)
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

    currentIds := make([]uint16, 3)
    for i, bt := range ib.Tags() {
        currentIds[i] = bt.tagId
    }

    if reflect.DeepEqual([]uint16 { 0x11, 0x22, 0x33 }, currentIds) == false {
        t.Fatalf("Pre-replace tags are not correct.")
    }

    bt = builderTag{
        tagId: 0xA9,
        value: "test string4",
    }

    err := ib.ReplaceAt(1, bt)
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
        tagId: 0x22,
        value: "test string3",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string4",
    }

    ib.Add(bt)


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


    err := ib.DeleteFirst(0x22)
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
        tagId: 0x22,
        value: "test string3",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string4",
    }

    ib.Add(bt)


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


    err := ib.DeleteN(0x22, 1)
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
        tagId: 0x22,
        value: "test string3",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string4",
    }

    ib.Add(bt)


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


    err := ib.DeleteN(0x22, 2)
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
        tagId: 0x22,
        value: "test string3",
    }

    ib.Add(bt)

    bt = builderTag{
        tagId: 0x33,
        value: "test string4",
    }

    ib.Add(bt)


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
