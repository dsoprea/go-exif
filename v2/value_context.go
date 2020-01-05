package exif

import (
	"encoding/binary"

	"github.com/dsoprea/go-exif/v2/common"
)

func newValueContextFromTag(ite *IfdTagEntry, addressableData []byte, byteOrder binary.ByteOrder) *exifcommon.ValueContext {
	return exifcommon.NewValueContext(
		ite.IfdPath,
		ite.TagId,
		ite.UnitCount,
		ite.ValueOffset,
		ite.RawValueOffset,
		addressableData,
		ite.TagType,
		byteOrder)
}
