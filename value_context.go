package exif

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
	UnitCount       uint32
	ValueOffset     uint32
	RawValueOffset  []byte
	AddressableData []byte
}

func newValueContext(unitCount, valueOffset uint32, rawValueOffset, addressableData []byte) ValueContext {
	return ValueContext{
		UnitCount:       unitCount,
		ValueOffset:     valueOffset,
		RawValueOffset:  rawValueOffset,
		AddressableData: addressableData,
	}
}
