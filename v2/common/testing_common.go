package exifcommon

import (
	"os"
	"path"

	"encoding/binary"
	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

var (
	assetsPath        = ""
	testImageFilepath = ""

	testExifData = make([]byte, 0)

	// EncodeDefaultByteOrder is the default byte-order for encoding operations.
	EncodeDefaultByteOrder = binary.BigEndian

	// Default byte order for tests.
	TestDefaultByteOrder = binary.BigEndian
)

func getModuleRootPath() string {
	currentWd, err := os.Getwd()
	log.PanicIf(err)

	currentPath := currentWd

	for {
		tryStampFilepath := path.Join(currentPath, ".MODULE_ROOT")

		f, err := os.Open(tryStampFilepath)
		if err == nil {
			f.Close()
			break
		}

		currentPath = path.Dir(currentPath)
		if currentPath == "/" {
			log.Panicf("could not find module-root")
		}
	}

	return currentPath
}

func init() {
	moduleRootPath := getModuleRootPath()
	assetsPath = path.Join(moduleRootPath, "assets")

	testImageFilepath = path.Join(assetsPath, "NDM_8901.jpg")

	// Load test EXIF data.

	filepath := path.Join(assetsPath, "NDM_8901.jpg.exif")

	var err error
	testExifData, err = ioutil.ReadFile(filepath)
	log.PanicIf(err)
}
