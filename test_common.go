package exif

import (
    "os"
    "path"

    "encoding/binary"
    "io/ioutil"

    "github.com/dsoprea/go-logging"
)

var (
    TestDefaultByteOrder = binary.BigEndian

    assetsPath = ""
    testExifData = make([]byte, 0)
)

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")


    // Load test EXIF data.

    filepath := path.Join(assetsPath, "NDM_8901.jpg.exif")

    exifData, err := ioutil.ReadFile(filepath)
    log.PanicIf(err)

// TODO(dustin): !! We're currently built to expect the JPEG EXIF header-prefix, but our test-data doesn't have that.So, artificially prefix it, for now.
    testExifData = make([]byte, len(exifData) + len(ExifHeaderPrefixBytes))
    copy(testExifData[0:], ExifHeaderPrefixBytes)
    copy(testExifData[len(ExifHeaderPrefixBytes):], exifData)
}
