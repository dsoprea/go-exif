package exif

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestAdd(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	originalBytes := []byte{0x11, 0x22, 0x33}

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x44,
		value:  NewIfdBuilderTagValueFromBytes([]byte(originalBytes)),
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

func TestSetNextIb(t *testing.T) {
	ib1 := NewIfdBuilder(RootIi, TestDefaultByteOrder)
	ib2 := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	if ib1.nextIb != nil {
		t.Fatalf("Next-IFD for IB1 not initially terminal.")
	}

	err := ib1.SetNextIb(ib2)
	log.PanicIf(err)

	if ib1.nextIb != ib2 {
		t.Fatalf("Next-IFD for IB1 not correct.")
	} else if ib2.nextIb != nil {
		t.Fatalf("Next-IFD for IB2 terminal.")
	}
}

func TestAddChildIb(t *testing.T) {

	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	exifIi, _ := IfdIdOrFail(IfdStandard, IfdExif)

	ibChild := NewIfdBuilder(exifIi, TestDefaultByteOrder)
	err = ib.AddChildIb(ibChild)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	if ib.tags[0].tagId != 0x11 {
		t.Fatalf("first tag not correct")
	} else if ib.tags[1].tagId != ibChild.ifdTagId {
		t.Fatalf("second tag ID does not match child-IFD tag-ID: (0x%04x) != (0x%04x)", ib.tags[1].tagId, ibChild.ifdTagId)
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

	entries := make([]*IfdTagEntry, 3)

	entries[0] = &IfdTagEntry{
		Ii:             ExifIi,
		TagId:          0x11,
		TagType:        TypeByte,
		UnitCount:      4,
		RawValueOffset: []byte{0x12, 0, 0, 0},
	}

	entries[1] = &IfdTagEntry{
		Ii:           ExifIi,
		TagId:        0x22,
		TagType:      TypeLong,
		ChildIfdName: "some ifd",
	}

	entries[2] = &IfdTagEntry{
		Ii:             ExifIi,
		TagId:          0x33,
		TagType:        TypeByte,
		UnitCount:      4,
		RawValueOffset: []byte{0x34, 0, 0, 0},
	}

	ifd := &Ifd{
		Ii:      RootIi,
		Entries: entries,
	}

	err := ib.AddTagsFromExisting(ifd, nil, nil, nil)
	log.PanicIf(err)

	if ib.tags[0].tagId != 0x11 {
		t.Fatalf("tag (0) not correct")
	} else if ib.tags[1].tagId != 0x22 {
		t.Fatalf("tag (1) not correct")
	} else if ib.tags[2].tagId != 0x33 {
		t.Fatalf("tag (2) not correct")
	} else if len(ib.tags) != 3 {
		t.Fatalf("tag count not correct")
	}
}

func TestAddTagsFromExisting__Includes(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	entries := make([]*IfdTagEntry, 3)

	entries[0] = &IfdTagEntry{
		Ii:      RootIi,
		TagType: TypeByte,
		TagId:   0x11,
	}

	entries[1] = &IfdTagEntry{
		Ii:           RootIi,
		TagType:      TypeByte,
		TagId:        0x22,
		ChildIfdName: "some ifd",
	}

	entries[2] = &IfdTagEntry{
		Ii:      RootIi,
		TagType: TypeByte,
		TagId:   0x33,
	}

	ifd := &Ifd{
		Ii:      RootIi,
		Entries: entries,
	}

	err := ib.AddTagsFromExisting(ifd, nil, []uint16{0x33}, nil)
	log.PanicIf(err)

	if ib.tags[0].tagId != 0x33 {
		t.Fatalf("tag (1) not correct")
	} else if len(ib.tags) != 1 {
		t.Fatalf("tag count not correct")
	}
}

func TestAddTagsFromExisting__Excludes(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	entries := make([]*IfdTagEntry, 3)

	entries[0] = &IfdTagEntry{
		Ii:      RootIi,
		TagType: TypeByte,
		TagId:   0x11,
	}

	entries[1] = &IfdTagEntry{
		Ii:           RootIi,
		TagType:      TypeByte,
		TagId:        0x22,
		ChildIfdName: "some ifd",
	}

	entries[2] = &IfdTagEntry{
		Ii:      RootIi,
		TagType: TypeByte,
		TagId:   0x33,
	}

	ifd := &Ifd{
		Ii:      RootIi,
		Entries: entries,
	}

	err := ib.AddTagsFromExisting(ifd, nil, nil, []uint16{0x11})
	log.PanicIf(err)

	if ib.tags[0].tagId != 0x22 {
		t.Fatalf("tag not correct")
	} else if len(ib.tags) != 2 {
		t.Fatalf("tag count not correct")
	}
}

func TestFindN_First_1(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
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
		log.Panicf("Found entry is not correct: (0x%04x)", bt.tagId)
	}
}

func TestFindN_First_2_1Returned(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
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
		log.Panicf("Found entry is not correct: (0x%04x)", bt.tagId)
	}
}

func TestFindN_First_2_2Returned(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string5")),
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
		log.Panicf("Found entry 0 is not correct: (0x%04x) [%s]", bt.tagId, bt.value)
	}

	bt = tags[found[1]]
	if bt.tagId != 0x11 || bytes.Compare(bt.value.Bytes(), []byte("test string4")) != 0 {
		log.Panicf("Found entry 1 is not correct: (0x%04x) [%s]", bt.tagId, bt.value)
	}
}

func TestFindN_Middle_WithDuplicates(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string5")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string6")),
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
		log.Panicf("Found entry is not correct: (0x%04x)", bt.tagId)
	}
}

func TestFindN_Middle_NoDuplicates(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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
		log.Panicf("Found entry is not correct: (0x%04x)", bt.tagId)
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

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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
		log.Panicf("Found entry is not correct: (0x%04x)", bt.tagId)
	}
}

func TestFind_Miss(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	currentIds := make([]uint16, 3)
	for i, bt := range ib.Tags() {
		currentIds[i] = bt.tagId
	}

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x33}, currentIds) == false {
		t.Fatalf("Pre-replace tags are not correct.")
	}

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x99,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
	}

	err = ib.Replace(0x22, bt)
	log.PanicIf(err)

	currentIds = make([]uint16, 3)
	for i, bt := range ib.Tags() {
		currentIds[i] = bt.tagId
	}

	if reflect.DeepEqual([]uint16{0x11, 0x99, 0x33}, currentIds) == false {
		t.Fatalf("Post-replace tags are not correct.")
	}
}

func TestReplaceN(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	currentIds := make([]uint16, 3)
	for i, bt := range ib.Tags() {
		currentIds[i] = bt.tagId
	}

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x33}, currentIds) == false {
		t.Fatalf("Pre-replace tags are not correct.")
	}

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0xA9,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
	}

	err = ib.ReplaceAt(1, bt)
	log.PanicIf(err)

	currentIds = make([]uint16, 3)
	for i, bt := range ib.Tags() {
		currentIds[i] = bt.tagId
	}

	if reflect.DeepEqual([]uint16{0x11, 0xA9, 0x33}, currentIds) == false {
		t.Fatalf("Post-replace tags are not correct.")
	}
}

func TestDeleteFirst(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x33}, currentIds) == false {
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

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x33}, currentIds) == false {
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

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x33}, currentIds) == false {
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

	bt := &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x11,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string")),
	}

	err := ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string2")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x22,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string3")),
	}

	err = ib.Add(bt)
	log.PanicIf(err)

	bt = &BuilderTag{
		ii:     RootIi,
		typeId: TypeByte,
		tagId:  0x33,
		value:  NewIfdBuilderTagValueFromBytes([]byte("test string4")),
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

	if reflect.DeepEqual([]uint16{0x11, 0x22, 0x22, 0x33}, currentIds) == false {
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

	if reflect.DeepEqual([]uint16{0x11, 0x33}, currentIds) == false {
		t.Fatalf("Post-delete tags not correct.")
	}

	err = ib.DeleteFirst(0x22)
	if err == nil {
		t.Fatalf("Expected an error.")
	} else if log.Is(err, ErrTagEntryNotFound) == false {
		log.Panic(err)
	}
}

func Test_IfdBuilder_CreateIfdBuilderFromExistingChain(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	filepath := path.Join(assetsPath, "NDM_8901.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	_, index, err := Collect(rawExif)
	log.PanicIf(err)

	itevr := NewIfdTagEntryValueResolver(rawExif, index.RootIfd.ByteOrder)
	ib := NewIfdBuilderFromExistingChain(index.RootIfd, itevr)

	actual := ib.DumpToStrings()

	expected := []string{
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(0) TAG=[0x010f]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(1) TAG=[0x0110]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(2) TAG=[0x0112]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(3) TAG=[0x011a]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(4) TAG=[0x011b]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(5) TAG=[0x0128]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(6) TAG=[0x0132]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(7) TAG=[0x013b]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(8) TAG=[0x0213]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[] INDEX=(9) TAG=[0x8298]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[Exif] INDEX=(10) TAG=[0x8769]",
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
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[Iop] INDEX=(25) TAG=[0xa005]",
		"<PARENTS=[IFD->Exif] IFD-NAME=[Iop]> IFD-TAG-ID=(0xa005) CHILD-IFD=[] INDEX=(0) TAG=[0x0001]",
		"<PARENTS=[IFD->Exif] IFD-NAME=[Iop]> IFD-TAG-ID=(0xa005) CHILD-IFD=[] INDEX=(1) TAG=[0x0002]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(26) TAG=[0xa20e]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(27) TAG=[0xa20f]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(28) TAG=[0xa210]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(29) TAG=[0xa401]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(30) TAG=[0xa402]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(31) TAG=[0xa403]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(32) TAG=[0xa406]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(33) TAG=[0xa430]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(34) TAG=[0xa431]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(35) TAG=[0xa432]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(36) TAG=[0xa434]",
		"<PARENTS=[IFD] IFD-NAME=[Exif]> IFD-TAG-ID=(0x8769) CHILD-IFD=[] INDEX=(37) TAG=[0xa435]",
		"<PARENTS=[] IFD-NAME=[IFD]> IFD-TAG-ID=(0x0000) CHILD-IFD=[GPSInfo] INDEX=(11) TAG=[0x8825]",
		"<PARENTS=[IFD] IFD-NAME=[GPSInfo]> IFD-TAG-ID=(0x8825) CHILD-IFD=[] INDEX=(0) TAG=[0x0000]",
	}

	if reflect.DeepEqual(actual, expected) == false {
		fmt.Printf("ACTUAL:\n%s\n\nEXPECTED:\n%s\n", strings.Join(actual, "\n"), strings.Join(expected, "\n"))
		t.Fatalf("IB did not [correctly] duplicate the IFD structure.")
	}
}

// TODO(dustin): !! Test with an actual GPS-attached image.

func Test_IfdBuilder_CreateIfdBuilderFromExistingChain_RealData(t *testing.T) {
	filepath := path.Join(assetsPath, "NDM_8901.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	// Decode from binary.

	_, originalIndex, err := Collect(rawExif)
	log.PanicIf(err)

	originalThumbnailData, err := originalIndex.RootIfd.NextIfd.Thumbnail()
	log.PanicIf(err)

	originalTags := originalIndex.RootIfd.DumpTags()

	// Encode back to binary.

	ibe := NewIfdByteEncoder()

	itevr := NewIfdTagEntryValueResolver(rawExif, originalIndex.RootIfd.ByteOrder)
	rootIb := NewIfdBuilderFromExistingChain(originalIndex.RootIfd, itevr)

	updatedExif, err := ibe.EncodeToExif(rootIb)
	log.PanicIf(err)

	// Parse again.

	_, recoveredIndex, err := Collect(updatedExif)
	log.PanicIf(err)

	recoveredTags := recoveredIndex.RootIfd.DumpTags()

	recoveredThumbnailData, err := recoveredIndex.RootIfd.NextIfd.Thumbnail()
	log.PanicIf(err)

	// Check the thumbnail.

	if bytes.Compare(recoveredThumbnailData, originalThumbnailData) != 0 {
		t.Fatalf("recovered thumbnail does not match original")
	}

	// Validate that all of the same IFDs were presented.

	originalIfdTags := make([][2]interface{}, 0)
	for _, ite := range originalTags {
		if ite.ChildIfdName != "" {
			originalIfdTags = append(originalIfdTags, [2]interface{}{ite.Ii, ite.TagId})
		}
	}

	recoveredIfdTags := make([][2]interface{}, 0)
	for _, ite := range recoveredTags {
		if ite.ChildIfdName != "" {
			recoveredIfdTags = append(recoveredIfdTags, [2]interface{}{ite.Ii, ite.TagId})
		}
	}

	if reflect.DeepEqual(recoveredIfdTags, originalIfdTags) != true {
		fmt.Printf("Original IFD tags:\n\n")

		for i, x := range originalIfdTags {
			fmt.Printf("  %02d %v\n", i, x)
		}

		fmt.Printf("\nRecovered IFD tags:\n\n")

		for i, x := range recoveredIfdTags {
			fmt.Printf("  %02d %v\n", i, x)
		}

		fmt.Printf("\n")

		t.Fatalf("Recovered IFD tags are not correct.")
	}

	// Validate that all of the tags owned by the IFDs were presented. Note
	// that the thumbnail tags are not kept but only produced on the fly, which
	// is why we check it above.

	if len(recoveredTags) != len(originalTags) {
		t.Fatalf("Recovered tag-count does not match original.")
	}

	for i, recoveredIte := range recoveredTags {
		if recoveredIte.ChildIfdName != "" {
			continue
		}

		originalIte := originalTags[i]

		if recoveredIte.Ii != originalIte.Ii {
			t.Fatalf("IfdIdentify not as expected: %s != %s  ITE=%s", recoveredIte.Ii, originalIte.Ii, recoveredIte)
		} else if recoveredIte.TagId != originalIte.TagId {
			t.Fatalf("Tag-ID not as expected: %d != %d  ITE=%s", recoveredIte.TagId, originalIte.TagId, recoveredIte)
		} else if recoveredIte.TagType != originalIte.TagType {
			t.Fatalf("Tag-type not as expected: %d != %d  ITE=%s", recoveredIte.TagType, originalIte.TagType, recoveredIte)
		}

		// TODO(dustin): We're always accessing the addressable-data using the root-IFD. It shouldn't matter, but we'd rather access it from our specific IFD.
		originalValueBytes, err := originalIte.ValueBytes(originalIndex.RootIfd.addressableData, originalIndex.RootIfd.ByteOrder)
		log.PanicIf(err)

		recoveredValueBytes, err := recoveredIte.ValueBytes(recoveredIndex.RootIfd.addressableData, recoveredIndex.RootIfd.ByteOrder)
		log.PanicIf(err)

		if bytes.Compare(originalValueBytes, recoveredValueBytes) != 0 {
			t.Fatalf("bytes of tag content not correct: %s != %s", originalIte, recoveredIte)
		}
	}
}

func Test_IfdBuilder_CreateIfdBuilderFromExistingChain_RealData_WithUpdate(t *testing.T) {
	filepath := path.Join(assetsPath, "NDM_8901.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	// Decode from binary.

	_, originalIndex, err := Collect(rawExif)
	log.PanicIf(err)

	originalThumbnailData, err := originalIndex.RootIfd.NextIfd.Thumbnail()
	log.PanicIf(err)

	originalTags := originalIndex.RootIfd.DumpTags()

	// Encode back to binary.

	ibe := NewIfdByteEncoder()

	itevr := NewIfdTagEntryValueResolver(rawExif, originalIndex.RootIfd.ByteOrder)
	rootIb := NewIfdBuilderFromExistingChain(originalIndex.RootIfd, itevr)

	// Update a tag,.

	exifBt, err := rootIb.FindTagWithName("ExifTag")
	log.PanicIf(err)

	ucBt, err := exifBt.value.Ib().FindTagWithName("UserComment")
	log.PanicIf(err)

	uc := TagUnknownType_9298_UserComment{
		EncodingType:  TagUnknownType_9298_UserComment_Encoding_ASCII,
		EncodingBytes: []byte("TEST COMMENT"),
	}

	err = ucBt.SetValue(rootIb.byteOrder, uc)
	log.PanicIf(err)

	// Encode.

	updatedExif, err := ibe.EncodeToExif(rootIb)
	log.PanicIf(err)

	// Parse again.

	_, recoveredIndex, err := Collect(updatedExif)
	log.PanicIf(err)

	recoveredTags := recoveredIndex.RootIfd.DumpTags()

	recoveredThumbnailData, err := recoveredIndex.RootIfd.NextIfd.Thumbnail()
	log.PanicIf(err)

	// Check the thumbnail.

	if bytes.Compare(recoveredThumbnailData, originalThumbnailData) != 0 {
		t.Fatalf("recovered thumbnail does not match original")
	}

	// Validate that all of the same IFDs were presented.

	originalIfdTags := make([][2]interface{}, 0)
	for _, ite := range originalTags {
		if ite.ChildIfdName != "" {
			originalIfdTags = append(originalIfdTags, [2]interface{}{ite.Ii, ite.TagId})
		}
	}

	recoveredIfdTags := make([][2]interface{}, 0)
	for _, ite := range recoveredTags {
		if ite.ChildIfdName != "" {
			recoveredIfdTags = append(recoveredIfdTags, [2]interface{}{ite.Ii, ite.TagId})
		}
	}

	if reflect.DeepEqual(recoveredIfdTags, originalIfdTags) != true {
		fmt.Printf("Original IFD tags:\n\n")

		for i, x := range originalIfdTags {
			fmt.Printf("  %02d %v\n", i, x)
		}

		fmt.Printf("\nRecovered IFD tags:\n\n")

		for i, x := range recoveredIfdTags {
			fmt.Printf("  %02d %v\n", i, x)
		}

		fmt.Printf("\n")

		t.Fatalf("Recovered IFD tags are not correct.")
	}

	// Validate that all of the tags owned by the IFDs were presented. Note
	// that the thumbnail tags are not kept but only produced on the fly, which
	// is why we check it above.

	if len(recoveredTags) != len(originalTags) {
		t.Fatalf("Recovered tag-count does not match original.")
	}

	for i, recoveredIte := range recoveredTags {
		if recoveredIte.ChildIfdName != "" {
			continue
		}

		originalIte := originalTags[i]

		if recoveredIte.Ii != originalIte.Ii {
			t.Fatalf("IfdIdentify not as expected: %s != %s  ITE=%s", recoveredIte.Ii, originalIte.Ii, recoveredIte)
		} else if recoveredIte.TagId != originalIte.TagId {
			t.Fatalf("Tag-ID not as expected: %d != %d  ITE=%s", recoveredIte.TagId, originalIte.TagId, recoveredIte)
		} else if recoveredIte.TagType != originalIte.TagType {
			t.Fatalf("Tag-type not as expected: %d != %d  ITE=%s", recoveredIte.TagType, originalIte.TagType, recoveredIte)
		}

		originalValueBytes, err := originalIte.ValueBytes(originalIndex.RootIfd.addressableData, originalIndex.RootIfd.ByteOrder)
		log.PanicIf(err)

		recoveredValueBytes, err := recoveredIte.ValueBytes(recoveredIndex.RootIfd.addressableData, recoveredIndex.RootIfd.ByteOrder)
		log.PanicIf(err)

		if recoveredIte.TagId == 0x9286 {
			expectedValueBytes := make([]byte, 0)

			expectedValueBytes = append(expectedValueBytes, []byte{'A', 'S', 'C', 'I', 'I', 0, 0, 0}...)
			expectedValueBytes = append(expectedValueBytes, []byte("TEST COMMENT")...)

			if bytes.Compare(recoveredValueBytes, expectedValueBytes) != 0 {
				t.Fatalf("Recovered UserComment does not have the right value: %v != %v", recoveredValueBytes, expectedValueBytes)
			}
		} else if bytes.Compare(recoveredValueBytes, originalValueBytes) != 0 {
			t.Fatalf("bytes of tag content not correct: %v != %v  ITE=%s", recoveredValueBytes, originalValueBytes, recoveredIte)
		}
	}
}

func ExampleIfd_Thumbnail() {
	filepath := path.Join(assetsPath, "NDM_8901.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	_, index, err := Collect(rawExif)
	log.PanicIf(err)

	thumbnailData, err := index.RootIfd.NextIfd.Thumbnail()
	log.PanicIf(err)

	thumbnailData = thumbnailData
	// Output:
}

func ExampleBuilderTag_SetValue() {
	filepath := path.Join(assetsPath, "NDM_8901.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	_, index, err := Collect(rawExif)
	log.PanicIf(err)

	// Create builder.

	itevr := NewIfdTagEntryValueResolver(rawExif, index.RootIfd.ByteOrder)
	rootIb := NewIfdBuilderFromExistingChain(index.RootIfd, itevr)

	// Find tag to update.

	exifBt, err := rootIb.FindTagWithName("ExifTag")
	log.PanicIf(err)

	ucBt, err := exifBt.value.Ib().FindTagWithName("UserComment")
	log.PanicIf(err)

	// Update the value. Since this is an "undefined"-type tag, we have to use
	// its type-specific struct.

	// TODO(dustin): !! Add an example for setting a non-unknown value, too.
	uc := TagUnknownType_9298_UserComment{
		EncodingType:  TagUnknownType_9298_UserComment_Encoding_ASCII,
		EncodingBytes: []byte("TEST COMMENT"),
	}

	err = ucBt.SetValue(rootIb.byteOrder, uc)
	log.PanicIf(err)

	// Encode.

	ibe := NewIfdByteEncoder()
	updatedExif, err := ibe.EncodeToExif(rootIb)
	log.PanicIf(err)

	updatedExif = updatedExif
	// Output:
}

func Test_IfdBuilder_CreateIfdBuilderWithExistingIfd(t *testing.T) {
	tagId := IfdTagIdWithIdentityOrFail(GpsIi)

	parentIfd := &Ifd{
		Ii:   RootIi,
		Name: IfdStandard,
	}

	ifd := &Ifd{
		Ii:        GpsIi,
		Name:      IfdGps,
		ByteOrder: TestDefaultByteOrder,
		Offset:    0x123,
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

func TestNewStandardBuilderTag_OneUnit(t *testing.T) {
	bt := NewStandardBuilderTag(ExifIi, uint16(0x8833), TestDefaultByteOrder, []uint32{uint32(0x1234)})

	if bt.ii != ExifIi {
		t.Fatalf("II in BuilderTag not correct")
	} else if bt.tagId != 0x8833 {
		t.Fatalf("tag-ID not correct")
	} else if bytes.Compare(bt.value.Bytes(), []byte{0x0, 0x0, 0x12, 0x34}) != 0 {
		t.Fatalf("value not correct")
	}
}

func TestNewStandardBuilderTag_TwoUnits(t *testing.T) {
	bt := NewStandardBuilderTag(ExifIi, uint16(0x8833), TestDefaultByteOrder, []uint32{uint32(0x1234), uint32(0x5678)})

	if bt.ii != ExifIi {
		t.Fatalf("II in BuilderTag not correct")
	} else if bt.tagId != 0x8833 {
		t.Fatalf("tag-ID not correct")
	} else if bytes.Compare(bt.value.Bytes(), []byte{
		0x0, 0x0, 0x12, 0x34,
		0x0, 0x0, 0x56, 0x78}) != 0 {
		t.Fatalf("value not correct")
	}
}

func TestNewStandardBuilderTagWithName(t *testing.T) {
	bt := NewStandardBuilderTagWithName(ExifIi, "ISOSpeed", TestDefaultByteOrder, []uint32{uint32(0x1234), uint32(0x5678)})

	if bt.ii != ExifIi {
		t.Fatalf("II in BuilderTag not correct")
	} else if bt.tagId != 0x8833 {
		t.Fatalf("tag-ID not correct")
	} else if bytes.Compare(bt.value.Bytes(), []byte{
		0x0, 0x0, 0x12, 0x34,
		0x0, 0x0, 0x56, 0x78}) != 0 {
		t.Fatalf("value not correct")
	}
}

func TestAddStandardWithName(t *testing.T) {
	ib := NewIfdBuilder(RootIi, TestDefaultByteOrder)

	err := ib.AddStandardWithName("ProcessingSoftware", "some software")
	log.PanicIf(err)

	if len(ib.tags) != 1 {
		t.Fatalf("Exactly one tag was not found: (%d)", len(ib.tags))
	}

	bt := ib.tags[0]

	if bt.ii != RootIi {
		t.Fatalf("II not correct: %s", bt.ii)
	} else if bt.tagId != 0x000b {
		t.Fatalf("Tag-ID not correct: (0x%04x)", bt.tagId)
	}

	s := string(bt.value.Bytes())

	if s != "some software\000" {
		t.Fatalf("Value not correct: (%d) [%s]", len(s), s)
	}
}
