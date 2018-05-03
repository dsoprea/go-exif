package exif

import (
    "testing"
)

func TestDumpBytesToString(t *testing.T) {
    s := DumpBytesToString([]byte { 0x12, 0x34, 0x56 })

    if s != "12 34 56" {
        t.Fatalf("result not expected")
    }
}

func TestDumpBytesClauseToString(t *testing.T) {
    s := DumpBytesClauseToString([]byte { 0x12, 0x34, 0x56 })

    if s != "0x12, 0x34, 0x56" {
        t.Fatalf("result not expected")
    }
}
