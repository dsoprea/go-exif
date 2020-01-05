package exifcommon

import (
	"testing"
	// "github.com/dsoprea/go-logging"
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
