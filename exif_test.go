package exif

import (
    "testing"
    "os"
    "path"
    "fmt"
    "reflect"

    "io/ioutil"

    "github.com/dsoprea/go-logging"
)

var (
    assetsPath = ""
)


func TestIsExif_True(t *testing.T) {
    e := NewExif()

    if ok := e.IsExif([]byte("Exif\000\000")); ok != true {
        t.Fatalf("expected true")
    }
}

func TestIsExif_False(t *testing.T) {
    e := NewExif()

    if ok := e.IsExif([]byte("something unexpected")); ok != false {
        t.Fatalf("expected false")
    }
}

func TestParse(t *testing.T) {
    // Open the file.

    filepath := path.Join(assetsPath, "NDM_8901.jpg")
    f, err := os.Open(filepath)
    log.PanicIf(err)

    defer f.Close()

    data, err := ioutil.ReadAll(f)
    log.PanicIf(err)

    // Search for the beginning of the EXIF information. The EXIF is near the
    // very beginning of our/most JPEGs, so this has a very low cost.

    e := NewExif()

    foundAt := -1
    for i := 0; i < len(data); i++ {
        if e.IsExif(data[i:i + 6]) == true {
            foundAt = i
            break
        }
    }

    if foundAt == -1 {
        log.Panicf("EXIF start not found")
    }

    // Run the parse.

    ti := NewTagIndex()
    tags := make([]string, 0)

    visitor := func(tagId uint16, tagType TagType, valueContext ValueContext) (err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
                log.PrintErrorf(err, "The visitor encountered an error.")
            }
        }()

        it, err := ti.GetWithTagId(tagId)
        if err != nil {
            if err == ErrTagNotFound {
                return nil
            } else {
                log.Panic(err)
            }
        }

        valueString, err := tagType.ValueString(valueContext, true)
        log.PanicIf(err)

        description := fmt.Sprintf("ID=(0x%04x) NAME=[%s] IFD=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]", tagId, it.Name, it.Ifd, valueContext.UnitCount, tagType.Name(), valueString)
        tags = append(tags, description)

        return nil
    }

    err = e.Parse(data[foundAt:], visitor)
    log.PanicIf(err)

    expected := []string {
        "ID=(0x010f) NAME=[Make] IFD=[Image] COUNT=(6) TYPE=[ASCII] VALUE=[Canon]",
        "ID=(0x0110) NAME=[Model] IFD=[Image] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]",
        "ID=(0x0112) NAME=[Orientation] IFD=[Image] COUNT=(1) TYPE=[SHORT] VALUE=[1]",
        "ID=(0x011a) NAME=[XResolution] IFD=[Image] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "ID=(0x011b) NAME=[YResolution] IFD=[Image] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "ID=(0x0128) NAME=[ResolutionUnit] IFD=[Image] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "ID=(0x0132) NAME=[DateTime] IFD=[Image] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]",
        "ID=(0x013b) NAME=[Artist] IFD=[Image] COUNT=(1) TYPE=[ASCII] VALUE=[]",
        "ID=(0x0213) NAME=[YCbCrPositioning] IFD=[Image] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "ID=(0x8298) NAME=[Copyright] IFD=[Image] COUNT=(1) TYPE=[ASCII] VALUE=[]",
        "ID=(0x8769) NAME=[ExifTag] IFD=[Image] COUNT=(1) TYPE=[LONG] VALUE=[360]",
        "ID=(0x8825) NAME=[GPSTag] IFD=[Image] COUNT=(1) TYPE=[LONG] VALUE=[9554]",
        "ID=(0x0103) NAME=[Compression] IFD=[Image] COUNT=(1) TYPE=[SHORT] VALUE=[6]",
        "ID=(0x011a) NAME=[XResolution] IFD=[Image] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "ID=(0x011b) NAME=[YResolution] IFD=[Image] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "ID=(0x0128) NAME=[ResolutionUnit] IFD=[Image] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "ID=(0x0201) NAME=[JPEGInterchangeFormat] IFD=[Image] COUNT=(1) TYPE=[LONG] VALUE=[11444]",
        "ID=(0x0202) NAME=[JPEGInterchangeFormatLength] IFD=[Image] COUNT=(1) TYPE=[LONG] VALUE=[21491]",
    }

    if reflect.DeepEqual(tags, expected) == false {
        t.Fatalf("tags not correct:\n%v", tags)
    }
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
}
