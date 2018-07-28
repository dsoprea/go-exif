package exif

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

	mi, err := im.Get([]uint16{IfdRootId, IfdExifId, IfdIopId})
	log.PanicIf(err)

	if mi.ParentTagId != IfdExifId {
		t.Fatalf("Parent tag-ID not correct")
	} else if mi.TagId != IfdIopId {
		t.Fatalf("Tag-ID not correct")
	} else if mi.Name != IfdIop {
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

	if mi.ParentTagId != IfdExifId {
		t.Fatalf("Parent tag-ID not correct")
	} else if mi.TagId != IfdIopId {
		t.Fatalf("Tag-ID not correct")
	} else if mi.Name != IfdIop {
		t.Fatalf("name not correct")
	} else if mi.PathPhrase() != "IFD/Exif/Iop" {
		t.Fatalf("path not correct")
	}
}
