package exif

import (
    "testing"

    "github.com/dsoprea/go-logging"
)

func TestGetWithTagId(t *testing.T) {
    ti := NewTagIndex()

    it, err := ti.GetWithTagId(0x10f)
    log.PanicIf(err)

    if it.Is(0x10f) == false || it.IsName("Image", "Make") == false {
        t.Fatalf("tag info not correct")
    }
}
