package exif

import (
    "fmt"
)

func DumpBytes(data []byte) {
    fmt.Printf("DUMP: ")
    for _, x := range data {
        fmt.Printf("%02x ", x)
    }

    fmt.Printf("\n")
}
