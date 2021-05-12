package exif

import (
	"os"
	"testing"

	"github.com/dsoprea/go-logging/v2"
	"github.com/dsoprea/go-utility/v2/filesystem"
)

func TestGpsDegreesEquals_Equals(t *testing.T) {
	gi := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	r := GpsDegreesEquals(gi, gi)
	if r != true {
		t.Fatalf("GpsDegrees structs were not equal as expected.")
	}
}

func TestGpsDegreesEquals_NotEqual_Orientation(t *testing.T) {
	gi1 := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	gi2 := gi1
	gi2.Orientation = 'B'

	r := GpsDegreesEquals(gi1, gi2)
	if r != false {
		t.Fatalf("GpsDegrees structs were equal but not supposed to be.")
	}
}

func TestGpsDegreesEquals_NotEqual_Position(t *testing.T) {
	gi1 := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	gi2 := gi1
	gi2.Minutes = 22.5

	r := GpsDegreesEquals(gi1, gi2)
	if r != false {
		t.Fatalf("GpsDegrees structs were equal but not supposed to be.")
	}
}

func TestGetFlatExifData(t *testing.T) {
	testExifData := getTestExifData()

	exifTags, _, err := GetFlatExifData(testExifData, nil)
	log.PanicIf(err)

	if len(exifTags) != 59 {
		t.Fatalf("Tag count not correct: (%d)", len(exifTags))
	}
}

func TestGetFlatExifDataUniversalSearch(t *testing.T) {
	testExifData := getTestExifData()

	exifTags, _, err := GetFlatExifDataUniversalSearch(testExifData, nil, false)
	log.PanicIf(err)

	if len(exifTags) != 59 {
		t.Fatalf("Tag count not correct: (%d)", len(exifTags))
	}
}

func TestGetFlatExifDataUniversalSearchWithReadSeeker(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	f, err := os.Open(testImageFilepath)
	log.PanicIf(err)

	defer f.Close()

	rawExif, err := SearchAndExtractExifWithReader(f)
	log.PanicIf(err)

	sb := rifs.NewSeekableBufferWithBytes(rawExif)

	exifTags, _, err := GetFlatExifDataUniversalSearchWithReadSeeker(sb, nil, false)
	log.PanicIf(err)

	if len(exifTags) != 59 {
		t.Fatalf("Tag count not correct: (%d)", len(exifTags))
	}
}
