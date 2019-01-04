package exif

import (
    "bytes"
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/dsoprea/go-logging"
)

func DumpBytes(data []byte) {
    fmt.Printf("DUMP: ")
    for _, x := range data {
        fmt.Printf("%02x ", x)
    }

    fmt.Printf("\n")
}

func DumpBytesClause(data []byte) {
    fmt.Printf("DUMP: ")

    fmt.Printf("[]byte { ")

    for i, x := range data {
        fmt.Printf("0x%02x", x)

        if i < len(data)-1 {
            fmt.Printf(", ")
        }
    }

    fmt.Printf(" }\n")
}

func DumpBytesToString(data []byte) string {
    b := new(bytes.Buffer)

    for i, x := range data {
        _, err := b.WriteString(fmt.Sprintf("%02x", x))
        log.PanicIf(err)

        if i < len(data)-1 {
            _, err := b.WriteRune(' ')
            log.PanicIf(err)
        }
    }

    return b.String()
}

func DumpBytesClauseToString(data []byte) string {
    b := new(bytes.Buffer)

    for i, x := range data {
        _, err := b.WriteString(fmt.Sprintf("0x%02x", x))
        log.PanicIf(err)

        if i < len(data)-1 {
            _, err := b.WriteString(", ")
            log.PanicIf(err)
        }
    }

    return b.String()
}

// ParseExifFullTimestamp parses dates like "2018:11:30 13:01:49".
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
