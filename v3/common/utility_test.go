package exifcommon

import (
	"fmt"
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
)

func TestDumpBytes(t *testing.T) {
	DumpBytes([]byte{1, 2, 3, 4})
}

func TestDumpBytesClause(t *testing.T) {
	DumpBytesClause([]byte{1, 2, 3, 4})
}

func TestDumpBytesToString(t *testing.T) {
	s := DumpBytesToString([]byte{1, 2, 3, 4})
	if s != "01 02 03 04" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestDumpBytesClauseToString(t *testing.T) {
	s := DumpBytesClauseToString([]byte{1, 2, 3, 4})
	if s != "0x01, 0x02, 0x03, 0x04" {
		t.Fatalf("Stringified clause is not correct: [%s]", s)
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

func ExampleExifFullTimestampString() {
	originalPhrase := "2018:11:30 13:01:49"

	timestamp, err := ParseExifFullTimestamp(originalPhrase)
	log.PanicIf(err)

	restoredPhrase := ExifFullTimestampString(timestamp)
	fmt.Printf("To EXIF timestamp: [%s]\n", restoredPhrase)

	// Output:
	// To EXIF timestamp: [2018:11:30 13:01:49]
}

func TestParseExifFullTimestamp(t *testing.T) {
	timestamp, err := ParseExifFullTimestamp("2018:11:30 13:01:49")
	log.PanicIf(err)

	actual := timestamp.Format(time.RFC3339)
	expected := "2018-11-30T13:01:49Z"

	if actual != expected {
		t.Fatalf("time not formatted correctly: [%s] != [%s]", actual, expected)
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
