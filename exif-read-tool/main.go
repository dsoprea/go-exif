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
        if e.IsExif(data[i:i + 6]) == true {
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
    visitor := func(indexedIfdName string, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
                log.Panic(err)
            }
        }()

        it, err := ti.Get(indexedIfdName, tagId)
        if err != nil {
            if log.Is(err, exif.ErrTagNotFound) {
                fmt.Printf("WARNING: Unknown tag: [%s] (%04x)\n", indexedIfdName, tagId)
                return nil
            } else {
                log.Panic(err)
            }
        }

        valueString := ""
        if tagType.Type() == exif.TypeUndefined {
            value, err := exif.UndefinedValue(indexedIfdName, tagId, valueContext, tagType.ByteOrder())
            if log.Is(err, exif.ErrUnhandledUnknownTypedTag) {
                valueString = "!UNDEFINED!"
            } else if err != nil {
                log.Panic(err)
            } else {
                valueString = fmt.Sprintf("%v", value)
            }
        } else {
            valueString, err = tagType.ValueString(valueContext, true)
            log.PanicIf(err)
        }

        fmt.Printf("IFD=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", indexedIfdName, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
        return nil
    }

    err = e.Parse(data[foundAt:], visitor)
    log.PanicIf(err)
}
