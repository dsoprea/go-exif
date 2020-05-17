package exif

import (
    "fmt"
    "math"
    "strconv"
    "strings"
    "time"

    "github.com/dsoprea/go-logging"

    "github.com/dsoprea/go-exif/v2/common"
    "github.com/dsoprea/go-exif/v2/undefined"
)

var (
    utilityLogger = log.NewLogger("exif.utility")
)

// ParseExifFullTimestamp parses dates like "2018:11:30 13:01:49" into a UTC
// `time.Time` struct.
func ParseExifFullTimestamp(fullTimestampPhrase string) (timestamp time.Time, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    parts := strings.Split(fullTimestampPhrase, " ")
    datestampValue, timestampValue := parts[0], parts[1]

    dateParts := strings.Split(datestampValue, ":")

    year, err := strconv.ParseUint(dateParts[0], 10, 16)
    if err != nil {
        log.Panicf("could not parse year")
    }

    month, err := strconv.ParseUint(dateParts[1], 10, 8)
    if err != nil {
        log.Panicf("could not parse month")
    }

    day, err := strconv.ParseUint(dateParts[2], 10, 8)
    if err != nil {
        log.Panicf("could not parse day")
    }

    timeParts := strings.Split(timestampValue, ":")

    hour, err := strconv.ParseUint(timeParts[0], 10, 8)
    if err != nil {
        log.Panicf("could not parse hour")
    }

    minute, err := strconv.ParseUint(timeParts[1], 10, 8)
    if err != nil {
        log.Panicf("could not parse minute")
    }

    second, err := strconv.ParseUint(timeParts[2], 10, 8)
    if err != nil {
        log.Panicf("could not parse second")
    }

    timestamp = time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC)
    return timestamp, nil
}

// ExifFullTimestampString produces a string like "2018:11:30 13:01:49" from a
// `time.Time` struct. It will attempt to convert to UTC first.
func ExifFullTimestampString(t time.Time) (fullTimestampPhrase string) {
    t = t.UTC()

    return fmt.Sprintf("%04d:%02d:%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

// ExifTag is one simple representation of a tag in a flat list of all of them.
type ExifTag struct {
    IfdPath string `json:"ifd_path"`

    TagId   uint16 `json:"id"`
    TagName string `json:"name"`

    TagTypeId   exifcommon.TagTypePrimitive `json:"type_id"`
    TagTypeName string                      `json:"type_name"`
    Value       interface{}                 `json:"value"`
    ValueBytes  []byte                      `json:"value_bytes"`

    FormattedFirst string `json:"formatted_first"`
    Formatted      string `json:"formatted"`

    ChildIfdPath string `json:"child_ifd_path"`
}

// String returns a string representation.
func (et ExifTag) String() string {
    return fmt.Sprintf("ExifTag<IFD-PATH=[%s] TAG-ID=(0x%02x) TAG-NAME=[%s] TAG-TYPE=[%s] VALUE=[%v] VALUE-BYTES=(%d) CHILD-IFD-PATH=[%s]", et.IfdPath, et.TagId, et.TagName, et.TagTypeName, et.FormattedFirst, len(et.ValueBytes), et.ChildIfdPath)
}

// GetFlatExifData returns a simple, flat representation of all tags.
func GetFlatExifData(exifData []byte) (exifTags []ExifTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    im := NewIfdMappingWithStandard()
    ti := NewTagIndex()

    _, index, err := Collect(im, ti, exifData)
    log.PanicIf(err)

    q := []*Ifd{index.RootIfd}

    exifTags = make([]ExifTag, 0)

    for len(q) > 0 {
        var ifd *Ifd
        ifd, q = q[0], q[1:]

        ti := NewTagIndex()
        for _, ite := range ifd.Entries {
            tagId := ite.TagId()
            tagType := ite.TagType()

            tagName := ""

            it, err := ti.Get(ifd.IfdPath, ite.tagId)
            if err != nil {
                if log.Is(err, ErrTagNotFound) != true {
                    log.Panic(err)
                }

                // This is an unknown tag.

                // This is supposed to be a convenience function and if we were
                // to keep the name empty or set it to some placeholder, it
                // might be mismanaged by the package that is calling us. If
                // they want to specifically manage these types of tags, they
                // can use more advanced functionality to specifically -handle
                // unknown tags.
                utilityLogger.Warningf(nil, "Tag with ID (0x%04x) in IFD [%s] is not recognized and will be ignored.", tagId, ifd.FqIfdPath)

                continue
            } else {
                tagName = it.Name
            }

            valueBytes, err := ite.GetRawBytes()
            if err != nil {
                if err == exifundefined.ErrUnparseableValue {
                    continue
                }

                log.Panic(err)
            }

            value, err := ite.Value()
            if err != nil {
                if err == exifcommon.ErrUnhandledUndefinedTypedTag {
                    value = exifundefined.UnparseableUnknownTagValuePlaceholder
                } else {
                    log.Panic(err)
                }
            }

            et := ExifTag{
                IfdPath:      ifd.FqIfdPath,
                TagId:        tagId,
                TagName:      tagName,
                TagTypeId:    tagType,
                TagTypeName:  tagType.String(),
                Value:        value,
                ValueBytes:   valueBytes,
                ChildIfdPath: ite.ChildIfdPath(),
            }

            et.Formatted, err = ite.Format()
            log.PanicIf(err)

            et.FormattedFirst, err = ite.FormatFirst()
            log.PanicIf(err)

            exifTags = append(exifTags, et)
        }

        for _, childIfd := range ifd.Children {
            q = append(q, childIfd)
        }

        if ifd.NextIfd != nil {
            q = append(q, ifd.NextIfd)
        }
    }

    return exifTags, nil
}

func GpsDegreesEquals(gi1, gi2 GpsDegrees) bool {
    if gi2.Orientation != gi1.Orientation {
        return false
    }

    degreesRightBound := math.Nextafter(gi1.Degrees, gi1.Degrees+1)
    minutesRightBound := math.Nextafter(gi1.Minutes, gi1.Minutes+1)
    secondsRightBound := math.Nextafter(gi1.Seconds, gi1.Seconds+1)

    if gi2.Degrees < gi1.Degrees || gi2.Degrees >= degreesRightBound {
        return false
    } else if gi2.Minutes < gi1.Minutes || gi2.Minutes >= minutesRightBound {
        return false
    } else if gi2.Seconds < gi1.Seconds || gi2.Seconds >= secondsRightBound {
        return false
    }

    return true
}

// FqIfdPath returns a fully-qualified IFD path expression.
func FqIfdPath(parentFqIfdName, ifdName string, ifdIndex int) string {
    var currentIfd string
    if parentFqIfdName != "" {
        currentIfd = fmt.Sprintf("%s/%s", parentFqIfdName, ifdName)
    } else {
        currentIfd = ifdName
    }

    if ifdIndex > 0 {
        currentIfd = fmt.Sprintf("%s%d", currentIfd, ifdIndex)
    }

    return currentIfd
}
