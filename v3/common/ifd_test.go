package exifcommon

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestIfdMapping_Add(t *testing.T) {
	im := NewIfdMapping()

	err := im.Add([]uint16{}, 0x1111, "ifd0")
	log.PanicIf(err)

	err = im.Add([]uint16{0x1111}, 0x4444, "ifd00")
	log.PanicIf(err)

	err = im.Add([]uint16{0x1111, 0x4444}, 0x5555, "ifd000")
	log.PanicIf(err)

	err = im.Add([]uint16{}, 0x2222, "ifd1")
	log.PanicIf(err)

	err = im.Add([]uint16{}, 0x3333, "ifd2")
	log.PanicIf(err)

	lineages, err := im.DumpLineages()
	log.PanicIf(err)

	sort.Strings(lineages)

	expected := []string{
		"ifd0",
		"ifd0/ifd00",
		"ifd0/ifd00/ifd000",
		"ifd1",
		"ifd2",
	}

	if reflect.DeepEqual(lineages, expected) != true {
		fmt.Printf("Actual:\n")
		fmt.Printf("\n")

		for i, line := range lineages {
			fmt.Printf("(%d) %s\n", i, line)
		}

		fmt.Printf("\n")

		fmt.Printf("Expected:\n")
		fmt.Printf("\n")

		for i, line := range expected {
			fmt.Printf("(%d) %s\n", i, line)
		}

		t.Fatalf("IFD-mapping dump not correct.")
	}
}

func TestIfdMapping_LoadStandardIfds(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	lineages, err := im.DumpLineages()
	log.PanicIf(err)

	sort.Strings(lineages)

	expected := []string{
		"IFD",
		"IFD/Exif",
		"IFD/Exif/Iop",
		"IFD/GPSInfo",
	}

	if reflect.DeepEqual(lineages, expected) != true {
		fmt.Printf("Actual:\n")
		fmt.Printf("\n")

		for i, line := range lineages {
			fmt.Printf("(%d) %s\n", i, line)
		}

		fmt.Printf("\n")

		fmt.Printf("Expected:\n")
		fmt.Printf("\n")

		for i, line := range expected {
			fmt.Printf("(%d) %s\n", i, line)
		}

		t.Fatalf("IFD-mapping dump not correct.")
	}
}

func TestIfdMapping_Get(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	mi, err := im.Get([]uint16{
		IfdStandardIfdIdentity.TagId(),
		IfdExifStandardIfdIdentity.TagId(),
		IfdExifIopStandardIfdIdentity.TagId(),
	})

	log.PanicIf(err)

	if mi.ParentTagId != IfdExifStandardIfdIdentity.TagId() {
		t.Fatalf("Parent tag-ID not correct")
	} else if mi.TagId != IfdExifIopStandardIfdIdentity.TagId() {
		t.Fatalf("Tag-ID not correct")
	} else if mi.Name != "Iop" {
		t.Fatalf("name not correct")
	} else if mi.PathPhrase() != "IFD/Exif/Iop" {
		t.Fatalf("path not correct")
	}
}

func TestIfdMapping_GetWithPath(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	mi, err := im.GetWithPath("IFD/Exif/Iop")
	log.PanicIf(err)

	if mi.ParentTagId != IfdExifStandardIfdIdentity.TagId() {
		t.Fatalf("Parent tag-ID not correct")
	} else if mi.TagId != IfdExifIopStandardIfdIdentity.TagId() {
		t.Fatalf("Tag-ID not correct")
	} else if mi.Name != "Iop" {
		t.Fatalf("name not correct")
	} else if mi.PathPhrase() != "IFD/Exif/Iop" {
		t.Fatalf("path not correct")
	}
}

func TestIfdMapping_ResolvePath__Regular(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	lineage, err := im.ResolvePath("IFD/Exif/Iop")
	log.PanicIf(err)

	expected := []IfdTagIdAndIndex{
		{Name: "IFD", TagId: 0, Index: 0},
		{Name: "Exif", TagId: 0x8769, Index: 0},
		{Name: "Iop", TagId: 0xa005, Index: 0},
	}

	if reflect.DeepEqual(lineage, expected) != true {
		t.Fatalf("Lineage not correct.")
	}
}

func TestIfdMapping_ResolvePath__WithIndices(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	lineage, err := im.ResolvePath("IFD/Exif1/Iop")
	log.PanicIf(err)

	expected := []IfdTagIdAndIndex{
		{Name: "IFD", TagId: 0, Index: 0},
		{Name: "Exif", TagId: 0x8769, Index: 1},
		{Name: "Iop", TagId: 0xa005, Index: 0},
	}

	if reflect.DeepEqual(lineage, expected) != true {
		t.Fatalf("Lineage not correct.")
	}
}

func TestIfdMapping_ResolvePath__Miss(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	_, err = im.ResolvePath("IFD/Exif/Invalid")
	if err == nil {
		t.Fatalf("Expected failure for invalid IFD path.")
	} else if err.Error() != "ifd child with name [Invalid] not registered: [IFD/Exif/Invalid]" {
		log.Panic(err)
	}
}

func TestIfdMapping_FqPathPhraseFromLineage(t *testing.T) {
	lineage := []IfdTagIdAndIndex{
		{Name: "IFD", Index: 0},
		{Name: "Exif", Index: 1},
		{Name: "Iop", Index: 0},
	}

	im := NewIfdMapping()

	fqPathPhrase := im.FqPathPhraseFromLineage(lineage)
	if fqPathPhrase != "IFD/Exif1/Iop" {
		t.Fatalf("path-phrase not correct: [%s]", fqPathPhrase)
	}
}

func TestIfdMapping_PathPhraseFromLineage(t *testing.T) {
	lineage := []IfdTagIdAndIndex{
		{Name: "IFD", Index: 0},
		{Name: "Exif", Index: 1},
		{Name: "Iop", Index: 0},
	}

	im := NewIfdMapping()

	fqPathPhrase := im.PathPhraseFromLineage(lineage)
	if fqPathPhrase != "IFD/Exif/Iop" {
		t.Fatalf("path-phrase not correct: [%s]", fqPathPhrase)
	}
}

func TestIfdMapping_NewIfdMappingWithStandard(t *testing.T) {
	imWith := NewIfdMappingWithStandard()
	imWithout := NewIfdMapping()

	err := LoadStandardIfds(imWithout)
	log.PanicIf(err)

	outputWith, err := imWith.DumpLineages()
	log.PanicIf(err)

	sort.Strings(outputWith)

	outputWithout, err := imWithout.DumpLineages()
	log.PanicIf(err)

	sort.Strings(outputWithout)

	if reflect.DeepEqual(outputWith, outputWithout) != true {
		fmt.Printf("WITH:\n")
		fmt.Printf("\n")

		for _, line := range outputWith {
			fmt.Printf("%s\n", line)
		}

		fmt.Printf("\n")

		fmt.Printf("WITHOUT:\n")
		fmt.Printf("\n")

		for _, line := range outputWithout {
			fmt.Printf("%s\n", line)
		}

		fmt.Printf("\n")

		t.Fatalf("Standard IFDs not loaded correctly.")
	}
}

func TestNewIfdIdentityFromString_Valid_WithoutIndexes(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	fqIfdPath := "IFD/Exif"

	ii, err := NewIfdIdentityFromString(im, fqIfdPath)
	log.PanicIf(err)

	if ii.String() != fqIfdPath {
		t.Fatalf("'%s' IFD-path was not parsed correctly: [%s]", fqIfdPath, ii.String())
	}
}

func TestNewIfdIdentityFromString_Valid_WithIndexes(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	fqIfdPath := "IFD2/Exif4"

	ii, err := NewIfdIdentityFromString(im, fqIfdPath)
	log.PanicIf(err)

	if ii.String() != fqIfdPath {
		t.Fatalf("'%s' IFD-path was not parsed correctly: [%s]", fqIfdPath, ii.String())
	}
}

func TestNewIfdIdentityFromString_Invalid_IfdPathJustRoot(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	fqIfdPath := "XYZ"

	_, err = NewIfdIdentityFromString(im, fqIfdPath)
	if err == nil {
		t.Fatalf("Expected error from invalid path.")
	} else if err.Error() != "ifd child with name [XYZ] not registered: [XYZ]" {
		log.Panic(err)
	}
}

func TestNewIfdIdentityFromString_Invalid_IfdPathWithSubdirectory(t *testing.T) {
	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	fqIfdPath := "IFD/XYZ"

	_, err = NewIfdIdentityFromString(im, fqIfdPath)
	if err == nil {
		t.Fatalf("Expected error from invalid path.")
	} else if err.Error() != "ifd child with name [XYZ] not registered: [IFD/XYZ]" {
		log.Panic(err)
	}
}
