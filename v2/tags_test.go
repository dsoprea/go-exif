package exif

import (
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGet(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.Get(IfdPathStandard, 0x10f)
	log.PanicIf(err)

	if it.Is(IfdPathStandard, 0x10f) == false || it.IsName(IfdPathStandard, "Make") == false {
		t.Fatalf("tag info not correct")
	}
}

func TestGetWithName(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.GetWithName(IfdPathStandard, "Make")
	log.PanicIf(err)

	if it.Is(IfdPathStandard, 0x10f) == false || it.Is(IfdPathStandard, 0x10f) == false {
		t.Fatalf("tag info not correct")
	}
}
