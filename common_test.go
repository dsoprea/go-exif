package exif

import (
	"os"
	"path"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

var (
	assetsPath   = ""
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

	var err error
	testExifData, err = ioutil.ReadFile(filepath)
	log.PanicIf(err)
}
