package exif

import (
	"testing"
)

func TestGpsDegreesEquals_Equals(t *testing.T) {
	gi := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	r := GpsDegreesEquals(gi, gi)
	if r != true {
		t.Fatalf("GpsDegrees structs were not equal as expected.")
	}
}

func TestGpsDegreesEquals_NotEqual_Orientation(t *testing.T) {
	gi1 := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	gi2 := gi1
	gi2.Orientation = 'B'

	r := GpsDegreesEquals(gi1, gi2)
	if r != false {
		t.Fatalf("GpsDegrees structs were equal but not supposed to be.")
	}
}

func TestGpsDegreesEquals_NotEqual_Position(t *testing.T) {
	gi1 := GpsDegrees{
		Orientation: 'A',
		Degrees:     11.0,
		Minutes:     22.0,
		Seconds:     33.0,
	}

	gi2 := gi1
	gi2.Minutes = 22.5

	r := GpsDegreesEquals(gi1, gi2)
	if r != false {
		t.Fatalf("GpsDegrees structs were equal but not supposed to be.")
	}
}
