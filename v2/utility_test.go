package exif

import (
    "fmt"
    "testing"
    "time"

    "github.com/dsoprea/go-logging"
)

func TestParseExifFullTimestamp(t *testing.T) {
    timestamp, err := ParseExifFullTimestamp("2018:11:30 13:01:49")
    log.PanicIf(err)

    actual := timestamp.Format(time.RFC3339)
    expected := "2018-11-30T13:01:49Z"

    if actual != expected {
        t.Fatalf("time not formatted correctly: [%s] != [%s]", actual, expected)
    }
}

func TestExifFullTimestampString(t *testing.T) {
    originalPhrase := "2018:11:30 13:01:49"

    timestamp, err := ParseExifFullTimestamp(originalPhrase)
    log.PanicIf(err)

    restoredPhrase := ExifFullTimestampString(timestamp)
    if restoredPhrase != originalPhrase {
        t.Fatalf("Final phrase [%s] does not equal original phrase [%s]", restoredPhrase, originalPhrase)
    }
}

func ExampleParseExifFullTimestamp() {
    originalPhrase := "2018:11:30 13:01:49"

    timestamp, err := ParseExifFullTimestamp(originalPhrase)
    log.PanicIf(err)

    fmt.Printf("To Go timestamp: [%s]\n", timestamp.Format(time.RFC3339))

    // Output:
    // To Go timestamp: [2018-11-30T13:01:49Z]
}

func ExampleExifFullTimestampString() {
    originalPhrase := "2018:11:30 13:01:49"

    timestamp, err := ParseExifFullTimestamp(originalPhrase)
    log.PanicIf(err)

    restoredPhrase := ExifFullTimestampString(timestamp)
    fmt.Printf("To EXIF timestamp: [%s]\n", restoredPhrase)

    // Output:
    // To EXIF timestamp: [2018:11:30 13:01:49]
}
