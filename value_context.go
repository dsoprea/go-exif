package exif

import (
	"encoding/binary"
)

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
	unitCount       uint32
	valueOffset     uint32
	rawValueOffset  []byte
	addressableData []byte

	tagType   TagTypePrimitive
	byteOrder binary.ByteOrder
}

func (vc ValueContext) UnitCount() uint32 {
	return vc.unitCount
}

func (vc ValueContext) ValueOffset() uint32 {
	return vc.valueOffset
}

func (vc ValueContext) RawValueOffset() []byte {
	return vc.rawValueOffset
}

func (vc ValueContext) AddressableData() []byte {
	return vc.addressableData
}

func newValueContext(unitCount, valueOffset uint32, rawValueOffset, addressableData []byte, tagType TagTypePrimitive, byteOrder binary.ByteOrder) ValueContext {
	return ValueContext{
		unitCount:       unitCount,
		valueOffset:     valueOffset,
		rawValueOffset:  rawValueOffset,
		addressableData: addressableData,

		tagType:   tagType,
		byteOrder: byteOrder,
	}
}
