package exif

import (
    "testing"
    "os"

    "io/ioutil"

    "github.com/dsoprea/go-logging"
)

func TestDumpBytes(t *testing.T) {
    f, err := ioutil.TempFile(os.TempDir(), "utilitytest")
    log.PanicIf(err)

    defer os.Remove(f.Name())

    originalStdout := os.Stdout
    os.Stdout = f

    DumpBytes([]byte { 0x11, 0x22 })

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

    DumpBytesClause([]byte { 0x11, 0x22 })

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
