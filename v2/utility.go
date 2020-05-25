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
    // IfdPath is the fully-qualified IFD path (even though it is not named as
    // such).
    IfdPath string `json:"ifd_path"`

    // TagId is the tag-ID.
    TagId uint16 `json:"id"`

    // TagName is the tag-name. This is never empty.
    TagName string `json:"name"`

    // UnitCount is the recorded number of units constution of the value.
    UnitCount uint32 `json:"unit_count"`

    // TagTypeId is the type-ID.
    TagTypeId exifcommon.TagTypePrimitive `json:"type_id"`

    // TagTypeName is the type name.
    TagTypeName string `json:"type_name"`

    // Value is the decoded value.
    Value interface{} `json:"value"`

    // ValueBytes is the raw, encoded value.
    ValueBytes []byte `json:"value_bytes"`

    // Formatted is the human representation of the first value (tag values are
    // always an array).
    FormattedFirst string `json:"formatted_first"`

    // Formatted is the human representation of the complete value.
    Formatted string `json:"formatted"`

    // ChildIfdPath is the name of the child IFD this tag represents (if it
    // represents any). Otherwise, this is empty.
    ChildIfdPath string `json:"child_ifd_path"`
}

// String returns a string representation.
func (et ExifTag) String() string {
    return fmt.Sprintf("ExifTag<IFD-PATH=[%s] TAG-ID=(0x%02x) TAG-NAME=[%s] TAG-TYPE=[%s] VALUE=[%v] VALUE-BYTES=(%d) CHILD-IFD-PATH=[%s]", et.IfdPath, et.TagId, et.TagName, et.TagTypeName, et.FormattedFirst, len(et.ValueBytes), et.ChildIfdPath)
}

// RELEASE(dustin): In the next release, make this return a list of skipped tags, too.

// GetFlatExifData returns a simple, flat representation of all tags.
func GetFlatExifData(exifData []byte) (exifTags []ExifTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    eh, err := ParseExifHeader(exifData)
    log.PanicIf(err)

    im := NewIfdMappingWithStandard()
    ti := NewTagIndex()

    ie := NewIfdEnumerate(im, ti, exifData, eh.ByteOrder)

    exifTags = make([]ExifTag, 0)

    visitor := func(fqIfdPath string, ifdIndex int, ite *IfdTagEntry) (err error) {
        tagId := ite.TagId()
        ii := ite.ifdIdentity

        it, err := ti.Get(ii, ite.tagId)
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
            utilityLogger.Warningf(nil, "Tag with ID (0x%04x) in IFD [%s] is not recognized and will be ignored.", tagId, fqIfdPath)

            it, err := ti.FindFirst(ite.tagId, nil)
            if err == nil {
                utilityLogger.Warningf(nil, "(cont'd) Tag [%s] with the same ID has been found in IFD [%s] and may be related. The tag you were looking for might have been written to the wrong IFD by a buggy implementation.", it.Name, it.IfdPath)
            }

            return nil
        }

        tagName := it.Name

        // This encodes down to base64. Since this an example tool and we do not
        // expect to ever decode the output, we are not worried about
        // specifically base64-encoding it in order to have a measure of
        // control.
        valueBytes, err := ite.GetRawBytes()
        if err != nil {
            if err == exifundefined.ErrUnparseableValue {
                return nil
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
            IfdPath:      fqIfdPath,
            TagId:        tagId,
            TagName:      tagName,
            UnitCount:    ite.UnitCount(),
            TagTypeId:    ite.TagType(),
            TagTypeName:  ite.TagType().String(),
            Value:        value,
            ValueBytes:   valueBytes,
            ChildIfdPath: ite.ChildIfdPath(),
        }

        et.Formatted, err = ite.Format()
        log.PanicIf(err)

        et.FormattedFirst, err = ite.FormatFirst()
        log.PanicIf(err)

        exifTags = append(exifTags, et)

        return nil
    }

    err = ie.Scan(exifcommon.IfdStandardIfdIdentity, eh.FirstIfdOffset, visitor)
    log.PanicIf(err)

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
