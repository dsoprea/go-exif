package exifcommon

import (
	"path"

	"go/build"

	"encoding/binary"
	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

var (
	assetsPath        = ""
	testImageFilepath = ""
	testExifData      = make([]byte, 0)
	moduleRootPath    = ""

	// EncodeDefaultByteOrder is the default byte-order for encoding operations.
	EncodeDefaultByteOrder = binary.BigEndian

	// Default byte order for tests.
	TestDefaultByteOrder = binary.BigEndian
)

// GetModuleRootPath returns our source-path when running from source during
// tests.
func GetModuleRootPath() string {
	p, err := build.Default.Import(
		"github.com/dsoprea/go-exif",
		build.Default.GOPATH,
		build.FindOnly)

	log.PanicIf(err)

	packagePath := p.Dir
	return path.Join(packagePath, "v2")
}

func getTestAssetsPath() string {
	moduleRootPath := GetModuleRootPath()
	assetsPath := path.Join(moduleRootPath, "assets")

	return assetsPath
}

func getTestImageFilepath() string {
	return path.Join(assetsPath, "NDM_8901.jpg")
}

func getTestExifData() []byte {
	filepath := path.Join(assetsPath, "NDM_8901.jpg.exif")

	var err error

	testExifData, err = ioutil.ReadFile(filepath)
	log.PanicIf(err)

	return testExifData
}

func init() {
	assetsPath = getTestAssetsPath()
}
