// This tool dumps EXIF information from images.
//
// Example command-line:
//
//   exif-read-tool -filepath <file-path>
//
// Example Output:
//
//   IFD=[IfdIdentity<PARENT-NAME=[] NAME=[IFD]>] ID=(0x010f) NAME=[Make] COUNT=(6) TYPE=[ASCII] VALUE=[Canon]
//   IFD=[IfdIdentity<PARENT-NAME=[] NAME=[IFD]>] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]
//   IFD=[IfdIdentity<PARENT-NAME=[] NAME=[IFD]>] ID=(0x0112) NAME=[Orientation] COUNT=(1) TYPE=[SHORT] VALUE=[1]
//   IFD=[IfdIdentity<PARENT-NAME=[] NAME=[IFD]>] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]
//   ...
package main

import (
	"flag"
	"fmt"
	"os"

	"encoding/json"
	"io/ioutil"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2"
	"github.com/dsoprea/go-exif/v2/common"
)

var (
	filepathArg     = ""
	printAsJsonArg  = false
	printLoggingArg = false
)

type IfdEntry struct {
	IfdPath     string                      `json:"ifd_path"`
	FqIfdPath   string                      `json:"fq_ifd_path"`
	IfdIndex    int                         `json:"ifd_index"`
	TagId       uint16                      `json:"tag_id"`
	TagName     string                      `json:"tag_name"`
	TagTypeId   exifcommon.TagTypePrimitive `json:"tag_type_id"`
	TagTypeName string                      `json:"tag_type_name"`
	UnitCount   uint32                      `json:"unit_count"`
	Value       interface{}                 `json:"value"`
	ValueString string                      `json:"value_string"`
}

type visitorFn func(fqIfdPath string, ifdIndex int, ite *exif.IfdTagEntry) (err error)

type visitorWrapper struct {
	visitor visitorFn
}

func (vw *visitorWrapper) Visit(fqIfdPath string, ifdIndex int, ite *exif.IfdTagEntry) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	err = vw.visitor(fqIfdPath, ifdIndex, ite)
	log.PanicIf(err)

	return nil
}

func main() {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.PrintErrorf(err, "Program error.")
			os.Exit(1)
		}
	}()

	flag.StringVar(&filepathArg, "filepath", "", "File-path of image")
	flag.BoolVar(&printAsJsonArg, "json", false, "Print JSON")
	flag.BoolVar(&printLoggingArg, "verbose", false, "Print logging")

	flag.Parse()

	if filepathArg == "" {
		fmt.Printf("Please provide a file-path for an image.\n")
		os.Exit(1)
	}

	if printLoggingArg == true {
		cla := log.NewConsoleLogAdapter()
		log.AddAdapter("console", cla)
	}

	f, err := os.Open(filepathArg)
	log.PanicIf(err)

	data, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	rawExif, err := exif.SearchAndExtractExif(data)
	log.PanicIf(err)

	// Run the parse.

	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	entries := make([]IfdEntry, 0)
	visitor := func(fqIfdPath string, ifdIndex int, ite *exif.IfdTagEntry) (err error) {
		defer func() {
			if state := recover(); state != nil {
				err = log.Wrap(state.(error))
				log.Panic(err)
			}
		}()

		tagId := ite.TagId()
		tagType := ite.TagType()

		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		log.PanicIf(err)

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if log.Is(err, exif.ErrTagNotFound) {
				fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ifdPath, tagId)
				return nil
			} else {
				log.Panic(err)
			}
		}

		value, err := ite.Value()
		log.PanicIf(err)

		valueString, err := ite.FormatFirst()
		log.PanicIf(err)

		entry := IfdEntry{
			IfdPath:     ifdPath,
			FqIfdPath:   fqIfdPath,
			IfdIndex:    ifdIndex,
			TagId:       tagId,
			TagName:     it.Name,
			TagTypeId:   tagType,
			TagTypeName: tagType.String(),
			UnitCount:   ite.UnitCount(),
			Value:       value,
			ValueString: valueString,
		}

		entries = append(entries, entry)

		return nil
	}

	vw := &visitorWrapper{visitor}

	_, err = exif.Visit(exifcommon.IfdStandard, im, ti, rawExif, vw)
	log.PanicIf(err)

	if printAsJsonArg == true {
		data, err := json.MarshalIndent(entries, "", "    ")
		log.PanicIf(err)

		fmt.Println(string(data))
	} else {
		for _, entry := range entries {
			fmt.Printf("IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", entry.IfdPath, entry.TagId, entry.TagName, entry.UnitCount, entry.TagTypeName, entry.ValueString)
		}
	}
}
