package exifundefined

import (
	"fmt"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
)

type Tag8828Oecf struct {
	Columns     uint16
	Rows        uint16
	ColumnNames []string
	Values      []exifcommon.SignedRational
}

func (oecf Tag8828Oecf) String() string {
	return fmt.Sprintf("Tag8828Oecf<COLUMNS=(%d) ROWS=(%d)>", oecf.Columns, oecf.Rows)
}

func (oecf Tag8828Oecf) EncoderName() string {
	return "Codec8828Oecf"
}

type Codec8828Oecf struct {
}

func (Codec8828Oecf) Decode(valueContext *exifcommon.ValueContext) (value EncodeableValue, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): Add test using known good data.

	valueContext.SetUndefinedValueType(exifcommon.TypeByte)

	valueBytes, err := valueContext.ReadBytes()
	log.PanicIf(err)

	oecf := Tag8828Oecf{}

	oecf.Columns = valueContext.ByteOrder().Uint16(valueBytes[0:2])
	oecf.Rows = valueContext.ByteOrder().Uint16(valueBytes[2:4])

	columnNames := make([]string, oecf.Columns)

	// startAt is where the current column name starts.
	startAt := 4

	// offset is our current position.
	offset := 4

	currentColumnNumber := uint16(0)

	for currentColumnNumber < oecf.Columns {
		if valueBytes[offset] == 0 {
			columnName := string(valueBytes[startAt:offset])
			if len(columnName) == 0 {
				log.Panicf("SFR column (%d) has zero length", currentColumnNumber)
			}

			columnNames[currentColumnNumber] = columnName
			currentColumnNumber++

			offset++
			startAt = offset
			continue
		}

		offset++
	}

	oecf.ColumnNames = columnNames

	rawRationalBytes := valueBytes[offset:]

	rationalSize := exifcommon.TypeSignedRational.Size()
	if len(rawRationalBytes)%rationalSize > 0 {
		log.Panicf("OECF signed-rationals not aligned: (%d) %% (%d) > 0", len(rawRationalBytes), rationalSize)
	}

	rationalCount := len(rawRationalBytes) / rationalSize

	parser := new(exifcommon.Parser)

	byteOrder := valueContext.ByteOrder()

	items, err := parser.ParseSignedRationals(rawRationalBytes, uint32(rationalCount), byteOrder)
	log.PanicIf(err)

	oecf.Values = items

	return oecf, nil
}

func init() {
	registerDecoder(
		exifcommon.IfdPathStandardExif,
		0x8828,
		Codec8828Oecf{})
}
