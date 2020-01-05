package exif

import (
	"fmt"

	"encoding/binary"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v2/common"
	"github.com/dsoprea/go-exif/v2/undefined"
)

var (
	iteLogger = log.NewLogger("exif.ifd_tag_entry")
)

type IfdTagEntry struct {
	TagId          uint16
	TagIndex       int
	TagType        exifcommon.TagTypePrimitive
	UnitCount      uint32
	ValueOffset    uint32
	RawValueOffset []byte

	// ChildIfdName is the right most atom in the IFD-path. We need this to
	// construct the fully-qualified IFD-path.
	ChildIfdName string

	// ChildIfdPath is the IFD-path of the child if this tag represents a child
	// IFD.
	ChildIfdPath string

	// ChildFqIfdPath is the IFD-path of the child if this tag represents a
	// child IFD. Includes indices.
	ChildFqIfdPath string

	// TODO(dustin): !! IB's host the child-IBs directly in the tag, but that's not the case here. Refactor to accomodate it for a consistent experience.

	// IfdPath is the IFD that this tag belongs to.
	IfdPath string

	// TODO(dustin): !! We now parse and read the value immediately. Update the rest of the logic to use this and get rid of all of the staggered and different resolution mechanisms.
	value              []byte
	isUnhandledUnknown bool
}

func (ite *IfdTagEntry) String() string {
	return fmt.Sprintf("IfdTagEntry<TAG-IFD-PATH=[%s] TAG-ID=(0x%04x) TAG-TYPE=[%s] UNIT-COUNT=(%d)>", ite.IfdPath, ite.TagId, ite.TagType.String(), ite.UnitCount)
}

// TODO(dustin): TODO(dustin): Stop exporting IfdPath and TagId.
//
// func (ite *IfdTagEntry) IfdPath() string {
// 	return ite.IfdPath
// }

// TODO(dustin): TODO(dustin): Stop exporting IfdPath and TagId.
//
// func (ite *IfdTagEntry) TagId() uint16 {
// 	return ite.TagId
// }

// ValueString renders a string from whatever the value in this tag is.
func (ite *IfdTagEntry) ValueString(addressableData []byte, byteOrder binary.ByteOrder) (phrase string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext :=
		newValueContextFromTag(
			ite,
			addressableData,
			byteOrder)

	if ite.TagType == exifcommon.TypeUndefined {
		var err error

		value, err := exifundefined.Decode(ite.IfdPath, ite.TagId, valueContext, byteOrder)
		log.PanicIf(err)

		s := value.(fmt.Stringer)

		phrase = s.String()
	} else {
		var err error

		phrase, err = valueContext.Format()
		log.PanicIf(err)
	}

	return phrase, nil
}

// ValueBytes renders a specific list of bytes from the value in this tag.
func (ite *IfdTagEntry) ValueBytes(addressableData []byte, byteOrder binary.ByteOrder) (value []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext :=
		newValueContextFromTag(
			ite,
			addressableData,
			byteOrder)

	// Return the exact bytes of the unknown-type value. Returning a string
	// (`ValueString`) is easy because we can just pass everything to
	// `Sprintf()`. Returning the raw, typed value (`Value`) is easy
	// (obviously). However, here, in order to produce the list of bytes, we
	// need to coerce whatever `Undefined()` returns.
	if ite.TagType == exifcommon.TypeUndefined {
		value, err := exifundefined.Decode(ite.IfdPath, ite.TagId, valueContext, byteOrder)
		log.PanicIf(err)

		ve := exifcommon.NewValueEncoder(byteOrder)

		ed, err := ve.Encode(value)
		log.PanicIf(err)

		return ed.Encoded, nil
	} else {
		rawBytes, err := valueContext.ReadRawEncoded()
		log.PanicIf(err)

		return rawBytes, nil
	}
}

// Value returns the specific, parsed, typed value from the tag.
func (ite *IfdTagEntry) Value(addressableData []byte, byteOrder binary.ByteOrder) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	valueContext :=
		newValueContextFromTag(
			ite,
			addressableData,
			byteOrder)

	if ite.TagType == exifcommon.TypeUndefined {
		var err error

		value, err = exifundefined.Decode(ite.IfdPath, ite.TagId, valueContext, byteOrder)
		log.PanicIf(err)
	} else {
		var err error

		value, err = valueContext.Values()
		log.PanicIf(err)
	}

	return value, nil
}
