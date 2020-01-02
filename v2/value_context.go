package exif

import (
	"encoding/binary"
)

func newValueContextFromTag(ite *IfdTagEntry, addressableData []byte, byteOrder binary.ByteOrder) *ValueContext {
	return newValueContext(
		ite.IfdPath,
		ite.TagId,
		ite.UnitCount,
		ite.ValueOffset,
		ite.RawValueOffset,
		addressableData,
		ite.TagType,
		byteOrder)
}
