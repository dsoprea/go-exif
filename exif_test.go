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

func TestVisit(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Exif failure.")
        }
    }()

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

    visitor := func(indexedIfdName string, tagId uint16, tagType TagType, valueContext ValueContext) (err error) {
        defer func() {
            if state := recover(); state != nil {
                err = log.Wrap(state.(error))
                log.Panic(err)
            }
        }()

        it, err := ti.Get(indexedIfdName, tagId)
        if err != nil {
            if log.Is(err, ErrTagNotFound) {
                fmt.Printf("Unknown tag: [%s] (%04x)\n", indexedIfdName, tagId)
                return nil
            } else {
                log.Panic(err)
            }
        }

        valueString := ""
        if tagType.Type() == TypeUndefined {
            value, err := UndefinedValue(indexedIfdName, tagId, valueContext, tagType.ByteOrder())
            if log.Is(err, ErrUnhandledUnknownTypedTag) {
                valueString = "!UNDEFINED!"
            } else if err != nil {
                log.Panic(err)
            } else {
                valueString = value.(string)
            }
        } else {
            valueString, err = tagType.ValueString(valueContext, true)
            log.PanicIf(err)
        }

        description := fmt.Sprintf("IFD=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]", indexedIfdName, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
        tags = append(tags, description)

        return nil
    }

    err = e.Visit(data[foundAt:], visitor)
    log.PanicIf(err)

    // for _, line := range tags {
    //     fmt.Printf("TAGS: %s\n", line)
    // }

    expected := []string {
        "IFD=[IFD] ID=(0x010f) NAME=[Make] COUNT=(6) TYPE=[ASCII] VALUE=[Canon]",
        "IFD=[IFD] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]",
        "IFD=[IFD] ID=(0x0112) NAME=[Orientation] COUNT=(1) TYPE=[SHORT] VALUE=[1]",
        "IFD=[IFD] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "IFD=[IFD] ID=(0x011b) NAME=[YResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "IFD=[IFD] ID=(0x0128) NAME=[ResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "IFD=[IFD] ID=(0x0132) NAME=[DateTime] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]",
        "IFD=[IFD] ID=(0x013b) NAME=[Artist] COUNT=(1) TYPE=[ASCII] VALUE=[]",
        "IFD=[IFD] ID=(0x0213) NAME=[YCbCrPositioning] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "IFD=[IFD] ID=(0x8298) NAME=[Copyright] COUNT=(1) TYPE=[ASCII] VALUE=[]",
        "IFD=[IFD] ID=(0x8769) NAME=[ExifTag] COUNT=(1) TYPE=[LONG] VALUE=[360]",
        "IFD=[Exif] ID=(0x829a) NAME=[ExposureTime] COUNT=(1) TYPE=[RATIONAL] VALUE=[1/640]",
        "IFD=[Exif] ID=(0x829d) NAME=[FNumber] COUNT=(1) TYPE=[RATIONAL] VALUE=[4/1]",
        "IFD=[Exif] ID=(0x8822) NAME=[ExposureProgram] COUNT=(1) TYPE=[SHORT] VALUE=[4]",
        "IFD=[Exif] ID=(0x8827) NAME=[ISOSpeedRatings] COUNT=(1) TYPE=[SHORT] VALUE=[1600]",
        "IFD=[Exif] ID=(0x8830) NAME=[SensitivityType] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "IFD=[Exif] ID=(0x8832) NAME=[RecommendedExposureIndex] COUNT=(1) TYPE=[LONG] VALUE=[1600]",
        "IFD=[Exif] ID=(0x9000) NAME=[ExifVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[0230]",
        "IFD=[Exif] ID=(0x9003) NAME=[DateTimeOriginal] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]",
        "IFD=[Exif] ID=(0x9004) NAME=[DateTimeDigitized] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]",
        "IFD=[Exif] ID=(0x9101) NAME=[ComponentsConfiguration] COUNT=(4) TYPE=[UNDEFINED] VALUE=[!UNDEFINED!]",
        "IFD=[Exif] ID=(0x9201) NAME=[ShutterSpeedValue] COUNT=(1) TYPE=[SRATIONAL] VALUE=[614400/65536]",
        "IFD=[Exif] ID=(0x9202) NAME=[ApertureValue] COUNT=(1) TYPE=[RATIONAL] VALUE=[262144/65536]",
        "IFD=[Exif] ID=(0x9204) NAME=[ExposureBiasValue] COUNT=(1) TYPE=[SRATIONAL] VALUE=[0/1]",
        "IFD=[Exif] ID=(0x9207) NAME=[MeteringMode] COUNT=(1) TYPE=[SHORT] VALUE=[5]",
        "IFD=[Exif] ID=(0x9209) NAME=[Flash] COUNT=(1) TYPE=[SHORT] VALUE=[16]",
        "IFD=[Exif] ID=(0x920a) NAME=[FocalLength] COUNT=(1) TYPE=[RATIONAL] VALUE=[16/1]",
        "IFD=[Exif] ID=(0x927c) NAME=[MakerNote] COUNT=(8152) TYPE=[UNDEFINED] VALUE=[!UNDEFINED!]",
        "IFD=[Exif] ID=(0x9286) NAME=[UserComment] COUNT=(264) TYPE=[UNDEFINED] VALUE=[!UNDEFINED!]",
        "IFD=[Exif] ID=(0x9290) NAME=[SubSecTime] COUNT=(3) TYPE=[ASCII] VALUE=[00]",
        "IFD=[Exif] ID=(0x9291) NAME=[SubSecTimeOriginal] COUNT=(3) TYPE=[ASCII] VALUE=[00]",
        "IFD=[Exif] ID=(0x9292) NAME=[SubSecTimeDigitized] COUNT=(3) TYPE=[ASCII] VALUE=[00]",
        "IFD=[Exif] ID=(0xa000) NAME=[FlashpixVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[0100]",
        "IFD=[Exif] ID=(0xa001) NAME=[ColorSpace] COUNT=(1) TYPE=[SHORT] VALUE=[1]",
        "IFD=[Exif] ID=(0xa002) NAME=[PixelXDimension] COUNT=(1) TYPE=[SHORT] VALUE=[3840]",
        "IFD=[Exif] ID=(0xa003) NAME=[PixelYDimension] COUNT=(1) TYPE=[SHORT] VALUE=[2560]",
        "IFD=[Exif] ID=(0xa005) NAME=[InteroperabilityTag] COUNT=(1) TYPE=[LONG] VALUE=[9326]",
        "IFD=[Iop] ID=(0x0001) NAME=[InteroperabilityIndex] COUNT=(4) TYPE=[ASCII] VALUE=[R98]",
        "IFD=[Iop] ID=(0x0002) NAME=[InteroperabilityVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[!UNDEFINED!]",
        "IFD=[Exif] ID=(0xa20e) NAME=[FocalPlaneXResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[3840000/1461]",
        "IFD=[Exif] ID=(0xa20f) NAME=[FocalPlaneYResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[2560000/972]",
        "IFD=[Exif] ID=(0xa210) NAME=[FocalPlaneResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "IFD=[Exif] ID=(0xa401) NAME=[CustomRendered] COUNT=(1) TYPE=[SHORT] VALUE=[0]",
        "IFD=[Exif] ID=(0xa402) NAME=[ExposureMode] COUNT=(1) TYPE=[SHORT] VALUE=[0]",
        "IFD=[Exif] ID=(0xa403) NAME=[WhiteBalance] COUNT=(1) TYPE=[SHORT] VALUE=[0]",
        "IFD=[Exif] ID=(0xa406) NAME=[SceneCaptureType] COUNT=(1) TYPE=[SHORT] VALUE=[0]",
        "IFD=[Exif] ID=(0xa430) NAME=[CameraOwnerName] COUNT=(1) TYPE=[ASCII] VALUE=[]",
        "IFD=[Exif] ID=(0xa431) NAME=[BodySerialNumber] COUNT=(13) TYPE=[ASCII] VALUE=[063024020097]",
        "IFD=[Exif] ID=(0xa432) NAME=[LensSpecification] COUNT=(4) TYPE=[RATIONAL] VALUE=[16/1]",
        "IFD=[Exif] ID=(0xa434) NAME=[LensModel] COUNT=(22) TYPE=[ASCII] VALUE=[EF16-35mm f/4L IS USM]",
        "IFD=[Exif] ID=(0xa435) NAME=[LensSerialNumber] COUNT=(11) TYPE=[ASCII] VALUE=[2400001068]",
        "IFD=[IFD] ID=(0x8825) NAME=[GPSTag] COUNT=(1) TYPE=[LONG] VALUE=[9554]",
        "IFD=[GPSInfo] ID=(0x0000) NAME=[GPSVersionID] COUNT=(4) TYPE=[BYTE] VALUE=[2]",
        "IFD=[IFD] ID=(0x0103) NAME=[Compression] COUNT=(1) TYPE=[SHORT] VALUE=[6]",
        "IFD=[IFD] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "IFD=[IFD] ID=(0x011b) NAME=[YResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]",
        "IFD=[IFD] ID=(0x0128) NAME=[ResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[2]",
        "IFD=[IFD] ID=(0x0201) NAME=[JPEGInterchangeFormat] COUNT=(1) TYPE=[LONG] VALUE=[11444]",
        "IFD=[IFD] ID=(0x0202) NAME=[JPEGInterchangeFormatLength] COUNT=(1) TYPE=[LONG] VALUE=[21491]",
    }

    if reflect.DeepEqual(tags, expected) == false {
        t.Fatalf("tags not correct:\n%v", tags)
    }
}

func TestCollect(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PrintErrorf(err, "Exif failure.")
        }
    }()

    e := NewExif()

    filepath := path.Join(assetsPath, "NDM_8901.jpg")

    rawExif, err := e.SearchAndExtractExif(filepath)
    log.PanicIf(err)

    rootIfd, tree, ifds, err := e.Collect(rawExif)
    log.PanicIf(err)

    rootIfd = rootIfd
    tree = tree
    ifds = ifds


// TODO(dustin): !! Finish test.


    // rootIfd.PrintTree()

    // expected := []string {
    // }

    // if reflect.DeepEqual(tags, expected) == false {
    //     t.Fatalf("tags not correct:\n%v", tags)
    // }
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath = path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
}
