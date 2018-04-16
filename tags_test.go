package exif

import (
    "testing"

    "github.com/dsoprea/go-logging"
)

func TestGet(t *testing.T) {
    ti := NewTagIndex()

    indexedIfdName := IfdName(IfdStandard, 0)

    it, err := ti.Get(indexedIfdName, 0x10f)
    log.PanicIf(err)

    if it.Is(0x10f) == false || it.IsName("IFD", "Make") == false {
        t.Fatalf("tag info not correct")
    }
}
