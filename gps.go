package exif

import (
    "fmt"
    "time"
)


type GpsDegrees struct {
    Orientation byte
    Degrees, Minutes, Seconds int
}

func (d GpsDegrees) String() string {
    return fmt.Sprintf("Degrees<O=[%s] D=(%d) M=(%d) S=(%d)>", string([]byte { d.Orientation }), d.Degrees, d.Minutes, d.Seconds)
}

func (d GpsDegrees) Decimal() float64 {
    decimal := float64(d.Degrees) + float64(d.Minutes) / 60.0 + float64(d.Seconds) / 3600.0

    if d.Orientation == 'S' || d.Orientation == 'W' {
        return -decimal
    } else {
        return decimal
    }
}


type GpsInfo struct {
    Latitude, Longitude GpsDegrees
    Altitude int
    Timestamp time.Time
}

func (gi GpsInfo) String() string {
    return fmt.Sprintf("GpsInfo<LAT=(%.05f) LON=(%.05f) ALT=(%d) TIME=[%s]>", gi.Latitude.Decimal(), gi.Longitude.Decimal(), gi.Altitude, gi.Timestamp)
}
