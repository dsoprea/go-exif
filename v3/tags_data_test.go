package exif

import (
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestGeotiffTags(t *testing.T) {
	testImageFilepath := getTestGeotiffFilepath()

	// Returns a slice starting with the EXIF data and going to the end of the
	// image.
	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	exifTags, _, err := GetFlatExifData(rawExif, nil)
	log.PanicIf(err)

	exifTagsIDMap := make(map[uint16]int)

	for _, e := range exifTags {
		exifTagsIDMap[e.TagId] = 1
	}

	if exifTagsIDMap[0x830e] == 0 {
		t.Fatal("Missing ModelPixelScaleTag.")
	}

	if exifTagsIDMap[0x8482] == 0 {
		t.Fatal("Missing ModelTiepointTag.")
	}
}
