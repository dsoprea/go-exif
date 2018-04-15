package exif

import (
    "testing"
    "os"
    "path"
    "fmt"

    "io/ioutil"

    "github.com/dsoprea/go-logging"
)

var (
    assetsPath = ""
)


func TestIsExif_True(t *testing.T) {
    e := NewExif()

    if ok := e.IsExif([]byte("Exif\000\000")); ok != true {
        t.Fatalf("expected true")
    }
}

func TestIsExif_False(t *testing.T) {
    e := NewExif()

    if ok := e.IsExif([]byte("something unexpected")); ok != false {
        t.Fatalf("expected false")
    }
}

func TestParse(t *testing.T) {
    // Open the file.

    filepath := path.Join(assetsPath, "NDM_8901.jpg")
    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    data, err := ioutil.ReadAll(f)
    log.PanicIf(err)

    // Search for the beginning of the EXIF information. The EXIF is near the
    // very beginning of our/most JPEGs, so this has a very low cost.

    e := NewExif()

    foundAt := -1
    for i := 0; i < len(data); i++ {
        if e.IsExif(data[i:i + 6]) == true {
            foundAt = i
            break
        }
    }

    if foundAt == -1 {
        log.Panicf("EXIF start not found")
    }

    // Run the parse.

    ti := NewTagIndex()

    visitor := func(tagId, tagType uint16, tagCount, valueOffset uint32) (err error) {
        it, err := ti.GetWithTagId(tagId)
        if err != nil {
            if err == ErrTagNotFound {
                return nil
            } else {
                log.Panic(err)
            }
        }

        fmt.Printf("Tag: ID=(0x%04x) NAME=[%s] IFD=[%s] TYPE=(%d) COUNT=(%d) VALUE-OFFSET=(%d)\n", tagId, it.Name, it.Ifd, tagType, tagCount, valueOffset)

// Notes on the tag-value's value (we'll have to use this as a pointer if the type potentially requires more than four bytes):
//
// This tag records the offset from the start of the TIFF header to the position where the value itself is
// recorded. In cases where the value fits in 4 Bytes, the value itself is recorded. If the value is smaller
// than 4 Bytes, the value is stored in the 4-Byte area starting from the left, i.e., from the lower end of
// the byte offset area. For example, in big endian format, if the type is SHORT and the value is 1, it is
// recorded as 00010000.H

        return nil
    }

    err = e.Parse(data[foundAt:], visitor)
    log.PanicIf(err)
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
}
