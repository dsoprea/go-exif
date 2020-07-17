package exif

import (
	"math"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestNewGpsDegreesFromRationals(t *testing.T) {
	latitudeRaw := []exifcommon.Rational{
		{Numerator: 22, Denominator: 2},
		{Numerator: 66, Denominator: 3},
		{Numerator: 132, Denominator: 4},
	}

	gd, err := NewGpsDegreesFromRationals("W", latitudeRaw)
	log.PanicIf(err)

	if gd.Orientation != 'W' {
		t.Fatalf("Orientation was not set correctly: [%s]", string([]byte{gd.Orientation}))
	}

	degreesRightBound := math.Nextafter(11.0, 12.0)
	minutesRightBound := math.Nextafter(22.0, 23.0)
	secondsRightBound := math.Nextafter(33.0, 34.0)

	if gd.Degrees < 11.0 || gd.Degrees >= degreesRightBound {
		t.Fatalf("Degrees is not correct: (%.2f)", gd.Degrees)
	} else if gd.Minutes < 22.0 || gd.Minutes >= minutesRightBound {
		t.Fatalf("Minutes is not correct: (%.2f)", gd.Minutes)
	} else if gd.Seconds < 33.0 || gd.Seconds >= secondsRightBound {
		t.Fatalf("Seconds is not correct: (%.2f)", gd.Seconds)
	}
}

func TestGpsDegrees_Raw(t *testing.T) {
	latitudeRaw := []exifcommon.Rational{
		{Numerator: 22, Denominator: 2},
		{Numerator: 66, Denominator: 3},
		{Numerator: 132, Denominator: 4},
	}

	gd, err := NewGpsDegreesFromRationals("W", latitudeRaw)
	log.PanicIf(err)

	actual := gd.Raw()

	expected := []exifcommon.Rational{
		{Numerator: 11, Denominator: 1},
		{Numerator: 22, Denominator: 1},
		{Numerator: 33, Denominator: 1},
	}

	if reflect.DeepEqual(actual, expected) != true {
		t.Fatalf("GpsInfo not correctly encoded down to raw: %v\n", actual)
	}
}
