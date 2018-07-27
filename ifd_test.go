package exif

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
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
	lines := make([]string, len(lineages))
	for i, lineage := range lineages {
		descriptions := make([]string, len(lineage))
		for i, mi := range lineage {
			descriptions[i] = fmt.Sprintf("(0x%04x) [%s]", mi.TagId, mi.Name)
		}

		lines[i] = strings.Join(descriptions, ", ")
	}

	sort.Strings(lines)

	expected := []string{
		"(0x1111) [ifd0]",
		"(0x1111) [ifd0], (0x4444) [ifd00]",
		"(0x1111) [ifd0], (0x4444) [ifd00], (0x5555) [ifd000]",
		"(0x2222) [ifd1]",
		"(0x3333) [ifd2]",
	}

	if reflect.DeepEqual(lines, expected) != true {
		fmt.Printf("Actual:\n")
		fmt.Printf("\n")

		for i, line := range lines {
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
	lines := make([]string, len(lineages))
	for i, lineage := range lineages {
		descriptions := make([]string, len(lineage))
		for i, mi := range lineage {
			descriptions[i] = fmt.Sprintf("(0x%04x) [%s]", mi.TagId, mi.Name)
		}

		lines[i] = strings.Join(descriptions, ", ")
	}

	sort.Strings(lines)

	expected := []string{
		"(0x0000) [IFD]",
		"(0x0000) [IFD], (0x8769) [Exif]",
		"(0x0000) [IFD], (0x8769) [Exif], (0xa005) [Iop]",
		"(0x0000) [IFD], (0x8825) [GPSInfo]",
	}

	if reflect.DeepEqual(lines, expected) != true {
		fmt.Printf("Actual:\n")
		fmt.Printf("\n")

		for i, line := range lines {
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
	}
}
