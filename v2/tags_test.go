package exif

import (
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

func TestIndexedTag_String(t *testing.T) {
	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TagTypePrimitive(11),
			exifcommon.TagTypePrimitive(22),
		},
	}

	if it.String() != "TAG<ID=(0x000b) NAME=[some_name] IFD=[ifd/path]>" {
		t.Fatalf("String output not correct: [%s]", it.String())
	}
}

func TestIndexedTag_IsName_True(t *testing.T) {
	it := &IndexedTag{
		Name:    "some_name",
		IfdPath: "ifd/path",
	}

	if it.IsName("ifd/path", "some_name") != true {
		t.Fatalf("IsName is not true.")
	}
}

func TestIndexedTag_IsName_FalseOnName(t *testing.T) {
	it := &IndexedTag{
		Name:    "some_name",
		IfdPath: "ifd/path",
	}

	if it.IsName("ifd/path", "some_name2") != false {
		t.Fatalf("IsName is not false.")
	}
}

func TestIndexedTag_IsName_FalseOnIfdPath(t *testing.T) {
	it := &IndexedTag{
		Name:    "some_name",
		IfdPath: "ifd/path",
	}

	if it.IsName("ifd/path2", "some_name") != false {
		t.Fatalf("IsName is not false.")
	}
}

func TestIndexedTag_Is_True(t *testing.T) {
	it := &IndexedTag{
		Id:      0x11,
		IfdPath: "ifd/path",
	}

	if it.Is("ifd/path", 0x11) != true {
		t.Fatalf("Is is not true.")
	}
}

func TestIndexedTag_Is_FalseOnId(t *testing.T) {
	it := &IndexedTag{
		Id:      0x11,
		IfdPath: "ifd/path",
	}

	if it.Is("ifd/path", 0x12) != false {
		t.Fatalf("Is is not false.")
	}
}

func TestIndexedTag_Is_FalseOnIfdName(t *testing.T) {
	it := &IndexedTag{
		Id:      0x11,
		IfdPath: "ifd/path",
	}

	if it.Is("ifd/path2", 0x11) != false {
		t.Fatalf("Is is not false.")
	}
}

func TestIndexedTag_GetEncodingType_WorksWithOneType(t *testing.T) {
	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeRational,
		},
	}

	if it.GetEncodingType(nil) != exifcommon.TypeRational {
		t.Fatalf("Expected the one type that was set.")
	}
}

func TestIndexedTag_GetEncodingType_FailsOnEmpty(t *testing.T) {
	// This also looks for an empty to reference the first spot, invalidly.

	defer func() {
		errRaw := recover()
		if errRaw == nil {
			t.Fatalf("Expected failure due to empty supported-types.")
		}

		err := errRaw.(error)
		if err.Error() != "IndexedTag [] (0) has no supported types." {
			log.Panic(err)
		}
	}()

	it := &IndexedTag{
		SupportedTypes: []exifcommon.TagTypePrimitive{},
	}

	it.GetEncodingType(nil)
}

func TestIndexedTag_GetEncodingType_PreferLongOverShort(t *testing.T) {
	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeShort,
		},
	}

	if it.GetEncodingType(nil) != exifcommon.TypeShort {
		t.Fatalf("Expected the second (LONG) type to be returned.")
	}

	it = &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeShort,
			exifcommon.TypeLong,
		},
	}

	if it.GetEncodingType(nil) != exifcommon.TypeLong {
		t.Fatalf("Expected the second (LONG) type to be returned.")
	}
}

func TestIndexedTag_GetEncodingType_BothRationalTypes(t *testing.T) {
	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeRational,
			exifcommon.TypeSignedRational,
		},
	}

	v1 := exifcommon.Rational{}

	if it.GetEncodingType(v1) != exifcommon.TypeRational {
		t.Fatalf("Expected the second (RATIONAL) type to be returned.")
	}

	v2 := exifcommon.SignedRational{}

	if it.GetEncodingType(v2) != exifcommon.TypeSignedRational {
		t.Fatalf("Expected the second (SIGNED RATIONAL) type to be returned.")
	}
}

func TestIndexedTag_DoesSupportType(t *testing.T) {
	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeRational,
			exifcommon.TypeSignedRational,
		},
	}

	if it.DoesSupportType(exifcommon.TypeRational) != true {
		t.Fatalf("Does not support unsigned-rational.")
	} else if it.DoesSupportType(exifcommon.TypeSignedRational) != true {
		t.Fatalf("Does not support signed-rational.")
	} else if it.DoesSupportType(exifcommon.TypeLong) != false {
		t.Fatalf("Does not support long.")
	}
}

func TestNewTagIndex(t *testing.T) {
	ti := NewTagIndex()

	if ti.tagsByIfd == nil {
		t.Fatalf("tagsByIfd is nil.")
	} else if ti.tagsByIfdR == nil {
		t.Fatalf("tagsByIfdR is nil.")
	}
}

func TestTagIndex_Add(t *testing.T) {
	ti := NewTagIndex()

	if len(ti.tagsByIfd) != 0 {
		t.Fatalf("tagsByIfd should be empty initially.")
	} else if len(ti.tagsByIfdR) != 0 {
		t.Fatalf("tagsByIfdR should be empty initially.")
	}

	it := &IndexedTag{
		Id:      0xb,
		Name:    "some_name",
		IfdPath: "ifd/path",
		SupportedTypes: []exifcommon.TagTypePrimitive{
			exifcommon.TypeRational,
			exifcommon.TypeSignedRational,
		},
	}

	err := ti.Add(it)
	log.PanicIf(err)

	if reflect.DeepEqual(ti.tagsByIfd[it.IfdPath][it.Id], it) != true {
		t.Fatalf("Not present in forward lookup.")
	} else if reflect.DeepEqual(ti.tagsByIfdR[it.IfdPath][it.Name], it) != true {
		t.Fatalf("Not present in reverse lookup.")
	}
}

func TestTagIndex_Get(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.Get(exifcommon.IfdStandardIfdIdentity, 0x10f)
	log.PanicIf(err)

	if it.Is(exifcommon.IfdStandardIfdIdentity.UnindexedString(), 0x10f) == false || it.IsName(exifcommon.IfdStandardIfdIdentity.UnindexedString(), "Make") == false {
		t.Fatalf("tag info not correct")
	}
}

func TestTagIndex_GetWithName(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.GetWithName(exifcommon.IfdStandardIfdIdentity, "Make")
	log.PanicIf(err)

	if it.Is(exifcommon.IfdStandardIfdIdentity.UnindexedString(), 0x10f) == false {
		t.Fatalf("tag info not correct")
	}
}

func TestTagIndex_FindFirst_HitOnFirst(t *testing.T) {

	searchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdExifStandardIfdIdentity,
		exifcommon.IfdStandardIfdIdentity,
	}

	ti := NewTagIndex()

	// ExifVersion
	it, err := ti.FindFirst(0x9000, searchOrder)
	log.PanicIf(err)

	if it.Is("IFD/Exif", 0x9000) != true {
		t.Fatalf("Returned tag is not correct.")
	}
}

func TestTagIndex_FindFirst_HitOnSecond(t *testing.T) {

	searchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdExifStandardIfdIdentity,
		exifcommon.IfdStandardIfdIdentity,
	}

	ti := NewTagIndex()

	// ProcessingSoftware
	it, err := ti.FindFirst(0x000b, searchOrder)
	log.PanicIf(err)

	if it.Is("IFD", 0x000b) != true {
		t.Fatalf("Returned tag is not correct.")
	}
}

func TestTagIndex_FindFirst_DefaultOrder_Miss(t *testing.T) {

	searchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdExifStandardIfdIdentity,
		exifcommon.IfdStandardIfdIdentity,
	}

	ti := NewTagIndex()

	_, err := ti.FindFirst(0x1234, searchOrder)
	if err == nil {
		t.Fatalf("Expected error for invalid tag.")
	} else if err != ErrTagNotFound {
		log.Panic(err)
	}
}

func TestTagIndex_FindFirst_ReverseDefaultOrder_HitOnSecond(t *testing.T) {

	reverseSearchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdStandardIfdIdentity,
		exifcommon.IfdExifStandardIfdIdentity,
	}

	ti := NewTagIndex()

	// ExifVersion
	it, err := ti.FindFirst(0x9000, reverseSearchOrder)
	log.PanicIf(err)

	if it.Is("IFD/Exif", 0x9000) != true {
		t.Fatalf("Returned tag is not correct.")
	}
}

func TestTagIndex_FindFirst_ReverseDefaultOrder_HitOnFirst(t *testing.T) {

	reverseSearchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdStandardIfdIdentity,
		exifcommon.IfdExifStandardIfdIdentity,
	}

	ti := NewTagIndex()

	// ProcessingSoftware
	it, err := ti.FindFirst(0x000b, reverseSearchOrder)
	log.PanicIf(err)

	if it.Is("IFD", 0x000b) != true {
		t.Fatalf("Returned tag is not correct.")
	}
}

func TestTagIndex_FindFirst_ReverseDefaultOrder_Miss(t *testing.T) {

	reverseSearchOrder := []*exifcommon.IfdIdentity{
		exifcommon.IfdStandardIfdIdentity,
		exifcommon.IfdExifStandardIfdIdentity,
	}

	ti := NewTagIndex()

	_, err := ti.FindFirst(0x1234, reverseSearchOrder)
	if err == nil {
		t.Fatalf("Expected error for invalid tag.")
	} else if err != ErrTagNotFound {
		log.Panic(err)
	}
}

func TestLoadStandardTags(t *testing.T) {
	ti := NewTagIndex()

	if len(ti.tagsByIfd) != 0 {
		t.Fatalf("tagsByIfd should be empty initially.")
	} else if len(ti.tagsByIfdR) != 0 {
		t.Fatalf("tagsByIfdR should be empty initially.")
	}

	err := LoadStandardTags(ti)
	log.PanicIf(err)

	if len(ti.tagsByIfd) == 0 {
		t.Fatalf("tagsByIfd should be non-empty at the end.")
	} else if len(ti.tagsByIfdR) == 0 {
		t.Fatalf("tagsByIfdR should be non-empty at the end.")
	}
}
