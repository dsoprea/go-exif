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
    "encoding/json"

    "github.com/dsoprea/go-logging"
    "github.com/dsoprea/go-exif"
)

var (
    filepathArg = ""
    printAsJsonArg = false
)


type IfdEntry struct {
    IfdName string `json:"ifd_name"`
    ParentIfdName string `json:"parent_ifd_name"`
    IfdIndex int `json:"ifd_index"`
    TagId uint16 `json:"tag_id"`
    TagName string `json:"tag_name"`
    TagTypeId uint16 `json:"tag_type_id"`
    TagTypeName string `json:"tag_type_name"`
    UnitCount uint32 `json:"unit_count"`
    Value interface{} `json:"value"`
    ValueString string `json:"value_string"`
}

func main() {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Program error.")
        }
    }()

    flag.StringVar(&filepathArg, "filepath", "", "File-path of image")
    flag.BoolVar(&printAsJsonArg, "json", false, "Print JSON")

    flag.Parse()

    if filepathArg == "" {
        fmt.Printf("Please provide a file-path for an image.\n")
        os.Exit(1)
    }

    f, err := os.Open(filepathArg)
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

    entries := make([]IfdEntry, 0)

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
        var value interface{}
        if tagType.Type() == exif.TypeUndefined {
            var err error
            value, err = exif.UndefinedValue(ii, tagId, valueContext, tagType.ByteOrder())
            if log.Is(err, exif.ErrUnhandledUnknownTypedTag) {
                value = nil
            } else if err != nil {
                log.Panic(err)
            } else {
                valueString = fmt.Sprintf("%v", value)
            }
        } else {
            valueString, err = tagType.ResolveAsString(valueContext, true)
            log.PanicIf(err)

            value = valueString
        }

        entry := IfdEntry{
            IfdName: ii.IfdName,
            ParentIfdName: ii.ParentIfdName,
            IfdIndex: ifdIndex,
            TagId: tagId,
            TagName: it.Name,
            TagTypeId: tagType.Type(),
            TagTypeName: tagType.Name(),
            UnitCount: valueContext.UnitCount,
            Value: value,
            ValueString: valueString,
        }

        entries = append(entries, entry)

        return nil
    }

    _, err = e.Visit(data[foundAt:], visitor)
    log.PanicIf(err)

    if printAsJsonArg == true {
        data, err := json.MarshalIndent(entries, "", "    ")
        log.PanicIf(err)

        fmt.Println(string(data))
    } else {
        for _, entry := range entries {
            fmt.Printf("IFD=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", entry.IfdName, entry.TagId, entry.TagName, entry.UnitCount, entry.TagTypeName, entry.ValueString)
        }
    }
}
