package exifcommon

import (
	"os"
	"path"

	"encoding/binary"

	log "github.com/dsoprea/go-logging"
)

var (
	moduleRootPath = ""

	// EncodeDefaultByteOrder is the default byte-order for encoding operations.
	EncodeDefaultByteOrder = binary.BigEndian

	// Default byte order for tests.
	TestDefaultByteOrder = binary.BigEndian
)

func GetModuleRootPath() string {
	if moduleRootPath == "" {
		moduleRootPath = os.Getenv("EXIF_MODULE_ROOT_PATH")
		if moduleRootPath != "" {
			return moduleRootPath
		}

		currentWd, err := os.Getwd()
		log.PanicIf(err)

		currentPath := currentWd

		visited := make([]string, 0)

		for {
			tryStampFilepath := path.Join(currentPath, ".MODULE_ROOT")

			_, err := os.Stat(tryStampFilepath)
			if err != nil && !os.IsNotExist(err) {
				log.Panic(err)
			} else if err == nil {
				break
			}

			visited = append(visited, tryStampFilepath)

			currentPath = path.Dir(currentPath)
			if currentPath == "/" {
				log.Panicf("could not find module-root: %v", visited)
			}
		}

		moduleRootPath = currentPath
	}

	return moduleRootPath
}

func GetTestAssetsPath() string {
	moduleRootPath := GetModuleRootPath()
	assetsPath := path.Join(moduleRootPath, "assets")

	return assetsPath
}
