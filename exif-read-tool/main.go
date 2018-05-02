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
    "os"
    "fmt"
    "flag"

    "io/ioutil"

    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-exif"
)

var (
    filepathArgument = ""
)

func main() {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Program error.")
        }
    }()

    flag.StringVar(&filepathArgument, "filepath", "", "File-path of image.")

    flag.Parse()

    if filepathArgument == "" {
        fmt.Printf("Please provide a file-path for an image.\n")
        os.Exit(1)
    }

    f, err := os.Open(filepathArgument)
    log.PanicIf(err)

    data, err := ioutil.ReadAll(f)
    log.PanicIf(err)

    e := exif.NewExif()

    foundAt := -1
    for i := 0; i < len(data); i++ {
        if exif.IsExif(data[i:i + 6]) == true {
            foundAt = i
            break
        }
    }

    if foundAt == -1 {
        fmt.Printf("EXIF data not found.\n")
        os.Exit(-1)
    }

    // Run the parse.

    ti := exif.NewTagIndex()
    visitor := func(ii exif.IfdIdentity, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
                log.Panic(err)
            }
        }()

        it, err := ti.Get(ii, tagId)
        if err != nil {
            if log.Is(err, exif.ErrTagNotFound) {
                fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", ii, tagId)
                return nil
            } else {
                log.Panic(err)
            }
        }

        valueString := ""
        if tagType.Type() == exif.TypeUndefined {
            value, err := exif.UndefinedValue(ii, tagId, valueContext, tagType.ByteOrder())
            if log.Is(err, exif.ErrUnhandledUnknownTypedTag) {
                valueString = "!UNDEFINED!"
            } else if err != nil {
                log.Panic(err)
            } else {
                valueString = fmt.Sprintf("%v", value)
            }
        } else {
            valueString, err = tagType.ResolveAsString(valueContext, true)
            log.PanicIf(err)
        }

        fmt.Printf("IFD=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", ii, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
        return nil
    }

    _, err = e.Visit(data[foundAt:], visitor)
    log.PanicIf(err)
}
