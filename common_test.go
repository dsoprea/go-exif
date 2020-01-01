package exif

import (
	"os"
	"path"
	"reflect"
	"testing"

	"io/ioutil"

	"github.com/dsoprea/go-logging"
)

var (
	assetsPath        = ""
	testImageFilepath = ""

	testExifData = make([]byte, 0)
)

func getExifSimpleTestIb() *IfdBuilder {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, IfdPathStandard, TestDefaultByteOrder)

	err = ib.AddStandard(0x000b, "asciivalue")
	log.PanicIf(err)

	err = ib.AddStandard(0x00ff, []uint16{0x1122})
	log.PanicIf(err)

	err = ib.AddStandard(0x0100, []uint32{0x33445566})
	log.PanicIf(err)

	err = ib.AddStandard(0x013e, []Rational{{Numerator: 0x11112222, Denominator: 0x33334444}})
	log.PanicIf(err)

	return ib
}

func getExifSimpleTestIbBytes() []byte {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	ti := NewTagIndex()
	ib := NewIfdBuilder(im, ti, IfdPathStandard, TestDefaultByteOrder)

	err = ib.AddStandard(0x000b, "asciivalue")
	log.PanicIf(err)

	err = ib.AddStandard(0x00ff, []uint16{0x1122})
	log.PanicIf(err)

	err = ib.AddStandard(0x0100, []uint32{0x33445566})
	log.PanicIf(err)

	err = ib.AddStandard(0x013e, []Rational{{Numerator: 0x11112222, Denominator: 0x33334444}})
	log.PanicIf(err)

	ibe := NewIfdByteEncoder()

	exifData, err := ibe.EncodeToExif(ib)
	log.PanicIf(err)

	return exifData
}

func validateExifSimpleTestIb(exifData []byte, t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	im := NewIfdMapping()

	err := LoadStandardIfds(im)
	log.PanicIf(err)

	ti := NewTagIndex()

	eh, index, err := Collect(im, ti, exifData)
	log.PanicIf(err)

	if eh.ByteOrder != TestDefaultByteOrder {
		t.Fatalf("EXIF byte-order is not correct: %v", eh.ByteOrder)
	} else if eh.FirstIfdOffset != ExifDefaultFirstIfdOffset {
		t.Fatalf("EXIF first IFD-offset not correct: (0x%02x)", eh.FirstIfdOffset)
	}

	if len(index.Ifds) != 1 {
		t.Fatalf("There wasn't exactly one IFD decoded: (%d)", len(index.Ifds))
	}

	ifd := index.RootIfd

	if ifd.ByteOrder != TestDefaultByteOrder {
		t.Fatalf("IFD byte-order not correct.")
	} else if ifd.IfdPath != IfdStandard {
		t.Fatalf("IFD name not correct.")
	} else if ifd.Index != 0 {
		t.Fatalf("IFD index not zero: (%d)", ifd.Index)
	} else if ifd.Offset != uint32(0x0008) {
		t.Fatalf("IFD offset not correct.")
	} else if len(ifd.Entries) != 4 {
		t.Fatalf("IFD number of entries not correct: (%d)", len(ifd.Entries))
	} else if ifd.NextIfdOffset != uint32(0) {
		t.Fatalf("Next-IFD offset is non-zero.")
	} else if ifd.NextIfd != nil {
		t.Fatalf("Next-IFD pointer is non-nil.")
	}

	// Verify the values by using the actual, orginal types (this is awesome).

	addressableData := exifData[ExifAddressableAreaStart:]

	expected := []struct {
		tagId uint16
		value interface{}
	}{
		{tagId: 0x000b, value: "asciivalue"},
		{tagId: 0x00ff, value: []uint16{0x1122}},
		{tagId: 0x0100, value: []uint32{0x33445566}},
		{tagId: 0x013e, value: []Rational{{Numerator: 0x11112222, Denominator: 0x33334444}}},
	}

	for i, e := range ifd.Entries {
		if e.TagId != expected[i].tagId {
			t.Fatalf("Tag-ID for entry (%d) not correct: (0x%02x) != (0x%02x)", i, e.TagId, expected[i].tagId)
		}

		value, err := e.Value(addressableData, TestDefaultByteOrder)
		log.PanicIf(err)

		if reflect.DeepEqual(value, expected[i].value) != true {
			t.Fatalf("Value for entry (%d) not correct: [%v] != [%v]", i, value, expected[i].value)
		}
	}
}

func init() {
	// This will only be executed when we're running tests in this package and
	// not when this package is being imported from a subpackage.

	goPath := os.Getenv("GOPATH")
	if goPath != "" {
		assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
	} else {
		// Module-enabled context.

		currentWd, err := os.Getwd()
		log.PanicIf(err)

		assetsPath = path.Join(currentWd, "assets")
	}

	testImageFilepath = path.Join(assetsPath, "NDM_8901.jpg")

	// Load test EXIF data.

	filepath := path.Join(assetsPath, "NDM_8901.jpg.exif")

	var err error
	testExifData, err = ioutil.ReadFile(filepath)
	log.PanicIf(err)
}
