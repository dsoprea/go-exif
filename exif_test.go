package exif

import (
    "testing"
    "os"
    "path"

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

    err = e.Parse(data[foundAt:])
    log.PanicIf(err)
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
}

