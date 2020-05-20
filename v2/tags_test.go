package exif

import (
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

func TestGet(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.Get(exifcommon.IfdStandardIfdIdentity.UnindexedString(), 0x10f)
	log.PanicIf(err)

	if it.Is(exifcommon.IfdStandardIfdIdentity.UnindexedString(), 0x10f) == false || it.IsName(exifcommon.IfdStandardIfdIdentity.UnindexedString(), "Make") == false {
		t.Fatalf("tag info not correct")
	}
}

func TestGetWithName(t *testing.T) {
	ti := NewTagIndex()

	it, err := ti.GetWithName(exifcommon.IfdStandardIfdIdentity.UnindexedString(), "Make")
	log.PanicIf(err)

	if it.Is(exifcommon.IfdStandardIfdIdentity.UnindexedString(), 0x10f) == false {
		t.Fatalf("tag info not correct")
	}
}
