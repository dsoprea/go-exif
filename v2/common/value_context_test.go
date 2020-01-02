package exifcommon

import (
	"bytes"
	"testing"

	"github.com/dsoprea/go-logging"
)

func TestValueContext_ReadAscii(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im := NewIfdMapping()

	err = LoadStandardIfds(im)
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifd := index.RootIfd

	var ite *IfdTagEntry
	for _, thisIte := range ifd.Entries {
		if thisIte.TagId == 0x0110 {
			ite = thisIte
			break
		}
	}

	if ite == nil {
		t.Fatalf("Tag not found.")
	}

	valueContext := ifd.GetValueContext(ite)

	decodedString, err := valueContext.ReadAscii()
	log.PanicIf(err)

	decodedBytes := []byte(decodedString)

	expected := []byte("Canon EOS 5D Mark III")

	if bytes.Compare(decodedBytes, expected) != 0 {
		t.Fatalf("Decoded bytes not correct.")
	}
}

func TestValueContext_Undefined(t *testing.T) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Test failure.")
		}
	}()

	rawExif, err := SearchFileAndExtractExif(testImageFilepath)
	log.PanicIf(err)

	im := NewIfdMapping()

	err = LoadStandardIfds(im)
	log.PanicIf(err)

	ti := NewTagIndex()

	_, index, err := Collect(im, ti, rawExif)
	log.PanicIf(err)

	ifdExif := index.Lookup[IfdPathStandardExif][0]

	var ite *IfdTagEntry
	for _, thisIte := range ifdExif.Entries {
		if thisIte.TagId == 0x9000 {
			ite = thisIte
			break
		}
	}

	if ite == nil {
		t.Fatalf("Tag not found.")
	}

	valueContext := ifdExif.GetValueContext(ite)

	value, err := valueContext.Undefined()
	log.PanicIf(err)

	gs, ok := value.(TagUnknownType_GeneralString)
	if ok != true {
		t.Fatalf("Undefined value not processed correctly.")
	}

	decodedBytes, err := gs.ValueBytes()
	log.PanicIf(err)

	expected := []byte("0230")

	if bytes.Compare(decodedBytes, expected) != 0 {
		t.Fatalf("Decoded bytes not correct.")
	}
}
