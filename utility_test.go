package exif

import (
	"io/ioutil"

	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dsoprea/go-logging"
)

func TestDumpBytes(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "utilitytest")
	log.PanicIf(err)

	defer os.Remove(f.Name())

	originalStdout := os.Stdout
	os.Stdout = f

	DumpBytes([]byte{0x11, 0x22})

	os.Stdout = originalStdout

	_, err = f.Seek(0, 0)
	log.PanicIf(err)

	content, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	if string(content) != "DUMP: 11 22 \n" {
		t.Fatalf("content not correct: [%s]", string(content))
	}
}

func TestDumpBytesClause(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "utilitytest")
	log.PanicIf(err)

	defer os.Remove(f.Name())

	originalStdout := os.Stdout
	os.Stdout = f

	DumpBytesClause([]byte{0x11, 0x22})

	os.Stdout = originalStdout

	_, err = f.Seek(0, 0)
	log.PanicIf(err)

	content, err := ioutil.ReadAll(f)
	log.PanicIf(err)

	if string(content) != "DUMP: []byte { 0x11, 0x22 }\n" {
		t.Fatalf("content not correct: [%s]", string(content))
	}
}

func TestDumpBytesToString(t *testing.T) {
	s := DumpBytesToString([]byte{0x12, 0x34, 0x56})

	if s != "12 34 56" {
		t.Fatalf("result not expected")
	}
}

func TestDumpBytesClauseToString(t *testing.T) {
	s := DumpBytesClauseToString([]byte{0x12, 0x34, 0x56})

	if s != "0x12, 0x34, 0x56" {
		t.Fatalf("result not expected")
	}
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
