package exif

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestIfdTagEntry_RawBytes_RealData(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	var ite *IfdTagEntry
	for _, thisIte := range index.RootIfd.entries {
		if thisIte.TagId() == 0x0110 {
			ite = thisIte
			break
		}
	}

	if ite == nil {
		t.Fatalf("Tag not found.")
	}

	decodedBytes, err := ite.GetRawBytes()
	log.PanicIf(err)

	expected := []byte("Canon EOS 5D Mark III")
	expected = append(expected, 0)

	if len(decodedBytes) != int(ite.UnitCount()) {
		t.Fatalf("Decoded bytes not the right count.")
	} else if bytes.Compare(decodedBytes, expected) != 0 {
		t.Fatalf("Decoded bytes not correct.")
	}
}

func TestIfd_FindTagWithId_Hit(t *testing.T) {
	testImageFilepath := getTestImageFilepath()
	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd
	results, err := ifd.FindTagWithId(0x011b)
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found: (%d)", len(results))
	} else if results[0].TagId() != 0x011b {
		t.Fatalf("The result was not expected: %v", results[0])
	}
}

func TestIfd_FindTagWithId_Miss(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	_, err = ifd.FindTagWithId(0xffff)
	if err == nil {
		t.Fatalf("Expected error for not-found tag.")
	} else if log.Is(err, ErrTagNotFound) == false {
		log.Panic(err)
	}
}

func TestIfd_FindTagWithName_Hit(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	results, err := ifd.FindTagWithName("YResolution")
	log.PanicIf(err)

	if len(results) != 1 {
		t.Fatalf("Exactly one result was not found: (%d)", len(results))
	} else if results[0].TagId() != 0x011b {
		t.Fatalf("The result was not expected: %v", results[0])
	}
}

func TestIfd_FindTagWithName_Miss(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	_, err = ifd.FindTagWithName("PlanarConfiguration")
	if err == nil {
		t.Fatalf("Expected error for not-found tag.")
	} else if log.Is(err, ErrTagNotFound) == false {
		log.Panic(err)
	}
}

func TestIfd_FindTagWithName_NonStandard(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	_, err = ifd.FindTagWithName("GeorgeNotAtHome")
	if err == nil {
		t.Fatalf("Expected error for not-found tag.")
	} else if log.Is(err, ErrTagNotKnown) == false {
		log.Panic(err)
	}
}

func TestIfd_Thumbnail(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	if ifd.nextIfd == nil {
		t.Fatalf("There is no IFD1.")
	}

	// The thumbnail is in IFD1 (The second root IFD).
	actual, err := ifd.nextIfd.Thumbnail()
	log.PanicIf(err)

	assetsPath := exifcommon.GetTestAssetsPath()
	expectedFilepath := path.Join(assetsPath, "NDM_8901.jpg.thumbnail")

	expected, err := ioutil.ReadFile(expectedFilepath)
	log.PanicIf(err)

	if bytes.Compare(actual, expected) != 0 {
		t.Fatalf("thumbnail not correct")
	}
}

func TestIfd_GpsInfo(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	filepath := getTestGpsImageFilepath()

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
	log.PanicIf(err)

	gi, err := ifd.GpsInfo()
	log.PanicIf(err)

	if gi.Latitude.Orientation != 'N' || gi.Latitude.Degrees != 26 || gi.Latitude.Minutes != 35 || gi.Latitude.Seconds != 12 {
		t.Fatalf("latitude not correct")
	} else if gi.Longitude.Orientation != 'W' || gi.Longitude.Degrees != 80 || gi.Longitude.Minutes != 3 || gi.Longitude.Seconds != 13 {
		t.Fatalf("longitude not correct")
	} else if gi.Altitude != 0 {
		t.Fatalf("altitude not correct")
	} else if gi.Timestamp.Unix() != 1524964977 {
		t.Fatalf("timestamp not correct")
	} else if gi.Altitude != 0 {
		t.Fatalf("altitude not correct")
	}
}

func TestIfd_GpsInfo__2_0_0_0(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	assetsPath := exifcommon.GetTestAssetsPath()
	filepath := path.Join(assetsPath, "gps-2000-scaled.jpg")

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
	log.PanicIf(err)

	gi, err := ifd.GpsInfo()
	log.PanicIf(err)

	expectedLatitude := GpsDegrees{
		Orientation: 'S',
		Degrees:     38.0,
		Minutes:     24.311687,
		Seconds:     0.0,
	}

	expectedLongitude := GpsDegrees{
		Orientation: 'E',
		Degrees:     144.0,
		Minutes:     11.33748,
		Seconds:     0.0,
	}

	if GpsDegreesEquals(gi.Latitude, expectedLatitude) != true {
		t.Fatalf("Latitude not correct: %v", gi.Latitude)
	} else if GpsDegreesEquals(gi.Longitude, expectedLongitude) != true {
		t.Fatalf("Longitude not correct: %v", gi.Longitude)
	} else if gi.Altitude != 0 {
		t.Fatalf("Altitude not correct: (%d)", gi.Altitude)
	} else if gi.Timestamp.Unix() != -62135596800 {
		t.Fatalf("Timestamp not correct: (%d)", gi.Timestamp.Unix())
	}
}

func TestIfd_EnumerateTagsRecursively(t *testing.T) {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	collected := make([][2]interface{}, 0)

	cb := func(ifd *Ifd, ite *IfdTagEntry) error {
		item := [2]interface{}{
			ifd.ifdIdentity.UnindexedString(),
			int(ite.TagId()),
		}

		collected = append(collected, item)

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	log.PanicIf(err)

	expected := [][2]interface{}{
		{"IFD", 0x010f},
		{"IFD", 0x0110},
		{"IFD", 0x0112},
		{"IFD", 0x011a},
		{"IFD", 0x011b},
		{"IFD", 0x0128},
		{"IFD", 0x0132},
		{"IFD", 0x013b},
		{"IFD", 0x0213},
		{"IFD", 0x8298},
		{"IFD/Exif", 0x829a},
		{"IFD/Exif", 0x829d},
		{"IFD/Exif", 0x8822},
		{"IFD/Exif", 0x8827},
		{"IFD/Exif", 0x8830},
		{"IFD/Exif", 0x8832},
		{"IFD/Exif", 0x9000},
		{"IFD/Exif", 0x9003},
		{"IFD/Exif", 0x9004},
		{"IFD/Exif", 0x9101},
		{"IFD/Exif", 0x9201},
		{"IFD/Exif", 0x9202},
		{"IFD/Exif", 0x9204},
		{"IFD/Exif", 0x9207},
		{"IFD/Exif", 0x9209},
		{"IFD/Exif", 0x920a},
		{"IFD/Exif", 0x927c},
		{"IFD/Exif", 0x9286},
		{"IFD/Exif", 0x9290},
		{"IFD/Exif", 0x9291},
		{"IFD/Exif", 0x9292},
		{"IFD/Exif", 0xa000},
		{"IFD/Exif", 0xa001},
		{"IFD/Exif", 0xa002},
		{"IFD/Exif", 0xa003},
		{"IFD/Exif/Iop", 0x0001},
		{"IFD/Exif/Iop", 0x0002},
		{"IFD/Exif", 0xa20e},
		{"IFD/Exif", 0xa20f},
		{"IFD/Exif", 0xa210},
		{"IFD/Exif", 0xa401},
		{"IFD/Exif", 0xa402},
		{"IFD/Exif", 0xa403},
		{"IFD/Exif", 0xa406},
		{"IFD/Exif", 0xa430},
		{"IFD/Exif", 0xa431},
		{"IFD/Exif", 0xa432},
		{"IFD/Exif", 0xa434},
		{"IFD/Exif", 0xa435},
		{"IFD/GPSInfo", 0x0000},
		{"IFD", 0x010f},
		{"IFD", 0x0110},
		{"IFD", 0x0112},
		{"IFD", 0x011a},
		{"IFD", 0x011b},
		{"IFD", 0x0128},
		{"IFD", 0x0132},
		{"IFD", 0x013b},
		{"IFD", 0x0213},
		{"IFD", 0x8298},
		{"IFD/Exif", 0x829a},
		{"IFD/Exif", 0x829d},
		{"IFD/Exif", 0x8822},
		{"IFD/Exif", 0x8827},
		{"IFD/Exif", 0x8830},
		{"IFD/Exif", 0x8832},
		{"IFD/Exif", 0x9000},
		{"IFD/Exif", 0x9003},
		{"IFD/Exif", 0x9004},
		{"IFD/Exif", 0x9101},
		{"IFD/Exif", 0x9201},
		{"IFD/Exif", 0x9202},
		{"IFD/Exif", 0x9204},
		{"IFD/Exif", 0x9207},
		{"IFD/Exif", 0x9209},
		{"IFD/Exif", 0x920a},
		{"IFD/Exif", 0x927c},
		{"IFD/Exif", 0x9286},
		{"IFD/Exif", 0x9290},
		{"IFD/Exif", 0x9291},
		{"IFD/Exif", 0x9292},
		{"IFD/Exif", 0xa000},
		{"IFD/Exif", 0xa001},
		{"IFD/Exif", 0xa002},
		{"IFD/Exif", 0xa003},
		{"IFD/Exif/Iop", 0x0001},
		{"IFD/Exif/Iop", 0x0002},
		{"IFD/Exif", 0xa20e},
		{"IFD/Exif", 0xa20f},
		{"IFD/Exif", 0xa210},
		{"IFD/Exif", 0xa401},
		{"IFD/Exif", 0xa402},
		{"IFD/Exif", 0xa403},
		{"IFD/Exif", 0xa406},
		{"IFD/Exif", 0xa430},
		{"IFD/Exif", 0xa431},
		{"IFD/Exif", 0xa432},
		{"IFD/Exif", 0xa434},
		{"IFD/Exif", 0xa435},
		{"IFD/GPSInfo", 0x0000},
	}

	if reflect.DeepEqual(collected, expected) != true {
		fmt.Printf("ACTUAL:\n")
		fmt.Printf("\n")

		for _, item := range collected {
			fmt.Printf("[2]interface{} { \"%s\", 0x%04x },\n", item[0], item[1])
		}

		fmt.Printf("\n")

		fmt.Printf("EXPECTED:\n")
		fmt.Printf("\n")

		for _, item := range expected {
			fmt.Printf("[2]interface{} { \"%s\", 0x%04x },\n", item[0], item[1])
		}

		fmt.Printf("\n")

		t.Fatalf("tags not visited correctly")
	}
}

func ExampleIfd_EnumerateTagsRecursively() {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	cb := func(ifd *Ifd, ite *IfdTagEntry) error {

		// Something useful.

		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)
	log.PanicIf(err)

	// Output:
}

func ExampleIfd_GpsInfo() {
	filepath := getTestGpsImageFilepath()

	rawExif, err := SearchFileAndExtractExif(filepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd, err := index.RootIfd.ChildWithIfdPath(exifcommon.IfdGpsInfoStandardIfdIdentity)
	log.PanicIf(err)

	gi, err := ifd.GpsInfo()
	log.PanicIf(err)

	fmt.Printf("%s\n", gi)

	// Output:
	// GpsInfo<LAT=(26.58667) LON=(-80.05361) ALT=(0) TIME=[2018-04-29 01:22:57 +0000 UTC]>
}

func ExampleIfd_FindTagWithName() {
	testImageFilepath := getTestImageFilepath()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im, err := exifcommon.NewIfdMappingWithStandard()
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	tagName := "Model"

	rootIfd := index.RootIfd

	// We know the tag we want is on IFD0 (the first/root IFD).
	results, err := rootIfd.FindTagWithName(tagName)
	log.PanicIf(err)

	// This should never happen.
	if len(results) != 1 {
		log.Panicf("there wasn't exactly one result")
	}

	ite := results[0]

	valueRaw, err := ite.Value()
	log.PanicIf(err)

	value := valueRaw.(string)
	fmt.Println(value)

	// Output:
	// Canon EOS 5D Mark III
}
