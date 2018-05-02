## Overview

This package provides native Go functionality to parse EXIF information out of images.


## *NOTICE*

This implementation is in active development. The reader functionality is reasonably complete and there are unit-tests covering the core functionality.

Remaining Tasks:

- A couple of tags needing special handling still need to be implemented. Currently these will show "!DEFINED!" for their values. See [here](type.go#L688) for more information
- Creating/updating tags.


## Getting

To get the project:

```
$ go get github.com/dsoprea/go-exif
```


## Testing

The traditional method:

```
$ go test github.com/dsoprea/go-exif
```


## Usage

Create an instance of the `Exif` type and call `Scan()` with a byte-slice, where the first byte is the beginning of the raw EXIF data. You may pass a callback that will be invoked for every tag or `nil` if you do not want one. If no callback is given, you are effectively just validating the structure or parsing of the image.

Obviously, it is most efficient to properly parse the media file and then provide the specific EXIF data to be parsed, but there is also a heuristic for finding the EXIF data within the media blob, directly. This means that, at least for testing or curiosity, **you do not have to parse or even understand the format of image or audio file in order to find and decode the EXIF information inside of it.** See the usage of the `IsExif` method in the example.


### Example Reader Tool

The example reader implementation below is included as a runnable tool:

```
$ go get github.com/dsoprea/go-exif/exif-read-tool
$ go build -o exif-read-tool github.com/dsoprea/go-exif/exif-read-tool
$ exif-read-tool -filepath "<media file-path>"
```

Example output:

```
IFD=[IFD] ID=(0x010f) NAME=[Make] COUNT=(6) TYPE=[ASCII] VALUE=[Canon]
IFD=[IFD] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]
IFD=[IFD] ID=(0x0112) NAME=[Orientation] COUNT=(1) TYPE=[SHORT] VALUE=[1]
IFD=[IFD] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]
IFD=[IFD] ID=(0x011b) NAME=[YResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[72/1]
IFD=[IFD] ID=(0x0128) NAME=[ResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[2]
IFD=[IFD] ID=(0x0132) NAME=[DateTime] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]
...
```


## Example

```
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
        valueString, err = tagType.ResolveAsString(valueContext, true)
        log.PanicIf(err)
    }

    fmt.Printf("IFD=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", indexedIfdName, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
    return nil
}

err = e.Visit(data[foundAt:], visitor)
log.PanicIf(err)
```


## *How You Can Help*

EXIF has an excellently-documented structure but there are a lot of devices and manufacturers out there. There are only so many files that we can personally find to test against, and most of these are images that have been generated only in the past few years. JPEG, being the largest implementor of EXIF, has been around for even longer (but not much). Therefore, there is a lot of different kinds of compatibility to test for.

**If you are able to help, it would be deeply appreciated if you could run the included reader-tool against all of the EXIF-compatible files you have. This is mostly going to be JPEG files (but not all variations). If you are able to test a large number of files (thousands or millions), please post an issue no matter what. Mention how many files you tried, whether there were any failures, and, if you would be willing, give us access to the failed files.**

If you are able to test 1M+ files, I will give you credit on the project. The further back in time your images reach, the higher in the list your name/company will go.
