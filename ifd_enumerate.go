package exif

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"encoding/binary"

	"github.com/dsoprea/go-logging"
)

var (
	ifdEnumerateLogger = log.NewLogger("exifjpeg.ifd")
)

var (
	ErrNoThumbnail = errors.New("no thumbnail")
	ErrNoGpsTags   = errors.New("no gps tags")
)

// IfdTagEnumerator knows how to decode an IFD and all of the tags it
// describes.
//
// The IFDs and the actual values can float throughout the EXIF block, but the
// IFD itself is just a minor header followed by a set of repeating,
// statically-sized records. So, the tags (though notnecessarily their values)
// are fairly simple to enumerate.
type IfdTagEnumerator struct {
	byteOrder       binary.ByteOrder
	addressableData []byte
	ifdOffset       uint32
	buffer          *bytes.Buffer
}

func NewIfdTagEnumerator(addressableData []byte, byteOrder binary.ByteOrder, ifdOffset uint32) (ite *IfdTagEnumerator) {
	ite = &IfdTagEnumerator{
		addressableData: addressableData,
		byteOrder:       byteOrder,
		buffer:          bytes.NewBuffer(addressableData[ifdOffset:]),
	}

	return ite
}

// getUint16 reads a uint16 and advances both our current and our current
// accumulator (which allows us to know how far to seek to the beginning of the
// next IFD when it's time to jump).
func (ife *IfdTagEnumerator) getUint16() (value uint16, raw []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	raw = make([]byte, 2)

	_, err = ife.buffer.Read(raw)
	log.PanicIf(err)

	if ife.byteOrder == binary.BigEndian {
		value = binary.BigEndian.Uint16(raw)
	} else {
		value = binary.LittleEndian.Uint16(raw)
	}

	return value, raw, nil
}

// getUint32 reads a uint32 and advances both our current and our current
// accumulator (which allows us to know how far to seek to the beginning of the
// next IFD when it's time to jump).
func (ife *IfdTagEnumerator) getUint32() (value uint32, raw []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	raw = make([]byte, 4)

	_, err = ife.buffer.Read(raw)
	log.PanicIf(err)

	if ife.byteOrder == binary.BigEndian {
		value = binary.BigEndian.Uint32(raw)
	} else {
		value = binary.LittleEndian.Uint32(raw)
	}

	return value, raw, nil
}

type IfdEnumerate struct {
	exifData      []byte
	buffer        *bytes.Buffer
	byteOrder     binary.ByteOrder
	currentOffset uint32
	tagIndex      *TagIndex
}

func NewIfdEnumerate(tagIndex *TagIndex, exifData []byte, byteOrder binary.ByteOrder) *IfdEnumerate {
	return &IfdEnumerate{
		exifData:  exifData,
		buffer:    bytes.NewBuffer(exifData),
		byteOrder: byteOrder,
		tagIndex:  tagIndex,
	}
}

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
	UnitCount       uint32
	ValueOffset     uint32
	RawValueOffset  []byte
	AddressableData []byte
}

func (ie *IfdEnumerate) getTagEnumerator(ifdOffset uint32) (ite *IfdTagEnumerator) {
	ite = NewIfdTagEnumerator(
		ie.exifData[ExifAddressableAreaStart:],
		ie.byteOrder,
		ifdOffset)

	return ite
}

func (ie *IfdEnumerate) parseTag(ii IfdIdentity, tagPosition int, ite *IfdTagEnumerator, resolveValue bool) (tag *IfdTagEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tagId, _, err := ite.getUint16()
	log.PanicIf(err)

	tagType, _, err := ite.getUint16()
	log.PanicIf(err)

	unitCount, _, err := ite.getUint32()
	log.PanicIf(err)

	valueOffset, rawValueOffset, err := ite.getUint32()
	log.PanicIf(err)

	tag = &IfdTagEntry{
		Ii:             ii,
		TagId:          tagId,
		TagIndex:       tagPosition,
		TagType:        tagType,
		UnitCount:      unitCount,
		ValueOffset:    valueOffset,
		RawValueOffset: rawValueOffset,
	}

	if resolveValue == true {
		value, isUnhandledUnknown, err := ie.resolveTagValue(tag)
		log.PanicIf(err)

		tag.value = value
		tag.isUnhandledUnknown = isUnhandledUnknown
	}

	// If it's an IFD but not a standard one, it'll just be seen as a LONG
	// (the standard IFD tag type), later, unless we skip it because it's
	// [likely] not even in the standard list of known tags.
	childIfdName, isIfd := IfdTagNameWithId(ii.IfdName, tagId)
	if isIfd == true {
		tag.ChildIfdName = childIfdName
	}

	return tag, nil
}

func (ie *IfdEnumerate) resolveTagValue(ite *IfdTagEntry) (valueBytes []byte, isUnhandledUnknown bool, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	addressableData := ie.exifData[ExifAddressableAreaStart:]

	// Return the exact bytes of the unknown-type value. Returning a string
	// (`ValueString`) is easy because we can just pass everything to
	// `Sprintf()`. Returning the raw, typed value (`Value`) is easy
	// (obviously). However, here, in order to produce the list of bytes, we
	// need to coerce whatever `UndefinedValue()` returns.
	if ite.TagType == TypeUndefined {
		valueContext := ValueContext{
			UnitCount:       ite.UnitCount,
			ValueOffset:     ite.ValueOffset,
			RawValueOffset:  ite.RawValueOffset,
			AddressableData: addressableData,
		}

		value, err := UndefinedValue(ite.Ii, ite.TagId, valueContext, ie.byteOrder)
		if err != nil {
			if log.Is(err, ErrUnhandledUnknownTypedTag) == true {
				valueBytes = []byte(UnparseableUnknownTagValuePlaceholder)
				return valueBytes, true, nil
			} else {
				log.Panic(err)
			}
		} else {
			switch value.(type) {
			case []byte:
				return value.([]byte), false, nil
			case string:
				return []byte(value.(string)), false, nil
			case UnknownTagValue:
				valueBytes, err := value.(UnknownTagValue).ValueBytes()
				log.PanicIf(err)

				return valueBytes, false, nil
			default:
				// TODO(dustin): !! Finish translating the rest of the types (make reusable and replace into other similar implementations?)
				log.Panicf("can not produce bytes for unknown-type tag (0x%04x): [%s]", ite.TagId, reflect.TypeOf(value))
			}
		}
	} else {
		originalType := NewTagType(ite.TagType, ie.byteOrder)
		byteCount := uint32(originalType.Size()) * ite.UnitCount

		tt := NewTagType(TypeByte, ie.byteOrder)

		if tt.ValueIsEmbedded(byteCount) == true {
			iteLogger.Debugf(nil, "Reading BYTE value (ITE; embedded).")

			// In this case, the bytes normally used for the offset are actually
			// data.
			valueBytes, err = tt.ParseBytes(ite.RawValueOffset, byteCount)
			log.PanicIf(err)
		} else {
			iteLogger.Debugf(nil, "Reading BYTE value (ITE; at offset).")

			valueBytes, err = tt.ParseBytes(addressableData[ite.ValueOffset:], byteCount)
			log.PanicIf(err)
		}
	}

	return valueBytes, false, nil
}

// RawTagVisitor is an optional callback that can get hit for every tag we parse
// through. `addressableData` is the byte array startign after the EXIF header
// (where the offsets of all IFDs and values are calculated from).
type RawTagVisitor func(ii IfdIdentity, ifdIndex int, tagId uint16, tagType TagType, valueContext ValueContext) (err error)

// ParseIfd decodes the IFD block that we're currently sitting on the first
// byte of.
func (ie *IfdEnumerate) ParseIfd(ii IfdIdentity, ifdIndex int, ite *IfdTagEnumerator, visitor RawTagVisitor, doDescend bool, resolveValues bool) (nextIfdOffset uint32, entries []*IfdTagEntry, thumbnailData []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tagCount, _, err := ite.getUint16()
	log.PanicIf(err)

	ifdEnumerateLogger.Debugf(nil, "Current IFD tag-count: (%d)", tagCount)

	entries = make([]*IfdTagEntry, 0)

	var iteThumbnailOffset *IfdTagEntry
	var iteThumbnailSize *IfdTagEntry

	for i := 0; i < int(tagCount); i++ {
		tag, err := ie.parseTag(ii, i, ite, resolveValues)
		log.PanicIf(err)

		if tag.TagId == ThumbnailOffsetTagId {
			iteThumbnailOffset = tag

			continue
		} else if tag.TagId == ThumbnailSizeTagId {
			iteThumbnailSize = tag
			continue
		}

		if visitor != nil {
			tt := NewTagType(tag.TagType, ie.byteOrder)

			vc := ValueContext{
				UnitCount:       tag.UnitCount,
				ValueOffset:     tag.ValueOffset,
				RawValueOffset:  tag.RawValueOffset,
				AddressableData: ie.exifData[ExifAddressableAreaStart:],
			}

			err := visitor(ii, ifdIndex, tag.TagId, tt, vc)
			log.PanicIf(err)
		}

		// If it's an IFD but not a standard one, it'll just be seen as a LONG
		// (the standard IFD tag type), later, unless we skip it because it's
		// [likely] not even in the standard list of known tags.
		if tag.ChildIfdName != "" && doDescend == true {
			ifdEnumerateLogger.Debugf(nil, "Descending to IFD [%s].", tag.ChildIfdName)

			childIi, _ := IfdIdOrFail(ii.IfdName, tag.ChildIfdName)

			err := ie.scan(childIi, tag.ValueOffset, visitor, resolveValues)
			log.PanicIf(err)
		}

		entries = append(entries, tag)
	}

	if iteThumbnailOffset != nil && iteThumbnailSize != nil {
		thumbnailData, err = ie.parseThumbnail(iteThumbnailOffset, iteThumbnailSize)
		log.PanicIf(err)
	}

	nextIfdOffset, _, err = ite.getUint32()
	log.PanicIf(err)

	ifdEnumerateLogger.Debugf(nil, "Next IFD at offset: (%08x)", nextIfdOffset)

	return nextIfdOffset, entries, thumbnailData, nil
}

func (ie *IfdEnumerate) parseThumbnail(offsetIte, lengthIte *IfdTagEntry) (thumbnailData []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	addressableData := ie.exifData[ExifAddressableAreaStart:]

	vRaw, err := lengthIte.Value(addressableData, ie.byteOrder)
	log.PanicIf(err)

	vList := vRaw.([]uint32)
	if len(vList) != 1 {
		log.Panicf("not exactly one long: (%d)", len(vList))
	}

	length := vList[0]

	// The tag is official a LONG type, but it's actually an offset to a blob of bytes.
	offsetIte.TagType = TypeByte
	offsetIte.UnitCount = length

	thumbnailData, err = offsetIte.ValueBytes(addressableData, ie.byteOrder)
	log.PanicIf(err)

	return thumbnailData, nil
}

// Scan enumerates the different EXIF's IFD blocks.
func (ie *IfdEnumerate) scan(ii IfdIdentity, ifdOffset uint32, visitor RawTagVisitor, resolveValues bool) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for ifdIndex := 0; ; ifdIndex++ {
		ifdEnumerateLogger.Debugf(nil, "Parsing IFD [%s] (%d) at offset (%04x).", ii.IfdName, ifdIndex, ifdOffset)
		ite := ie.getTagEnumerator(ifdOffset)

		nextIfdOffset, _, _, err := ie.ParseIfd(ii, ifdIndex, ite, visitor, true, resolveValues)
		log.PanicIf(err)

		if nextIfdOffset == 0 {
			break
		}

		ifdOffset = nextIfdOffset
	}

	return nil
}

// Scan enumerates the different EXIF blocks (called IFDs).
func (ie *IfdEnumerate) Scan(ifdOffset uint32, visitor RawTagVisitor, resolveValue bool) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ii, _ := IfdIdOrFail("", IfdStandard)

	err = ie.scan(ii, ifdOffset, visitor, resolveValue)
	log.PanicIf(err)

	return nil
}

// Ifd represents a single parsed IFD.
type Ifd struct {
	// This is just for convenience, just so that we can easily get the values
	// and not involve other projects in semantics that they won't otherwise
	// need to know.
	addressableData []byte

	ByteOrder binary.ByteOrder

	Ii    IfdIdentity
	TagId uint16

	Id int

	ParentIfd *Ifd

	// ParentTagIndex is our tag position in the parent IFD, if we had a parent
	// (if `ParentIfd` is not nil and we weren't an IFD referenced as a sibling
	// instead of as a child).
	ParentTagIndex int

	Name   string
	Index  int
	Offset uint32

	Entries        []*IfdTagEntry
	EntriesByTagId map[uint16][]*IfdTagEntry

	Children []*Ifd

	ChildIfdIndex map[string]*Ifd

	NextIfdOffset uint32
	NextIfd       *Ifd

	thumbnailData []byte

	tagIndex *TagIndex
}

func (ifd *Ifd) ChildWithIfdIdentity(ii IfdIdentity) (childIfd *Ifd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for _, childIfd := range ifd.Children {
		if childIfd.Ii == ii {
			return childIfd, nil
		}
	}

	log.Panic(ErrTagNotFound)
	return nil, nil
}

func (ifd *Ifd) ChildWithName(ifdName string) (childIfd *Ifd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for _, childIfd := range ifd.Children {
		if childIfd.Ii.IfdName == ifdName {
			return childIfd, nil
		}
	}

	log.Panic(ErrTagNotFound)
	return nil, nil
}

func (ifd *Ifd) TagValue(ite *IfdTagEntry) (value interface{}, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	value, err = ite.Value(ifd.addressableData, ifd.ByteOrder)
	log.PanicIf(err)

	return value, nil
}

func (ifd *Ifd) TagValueBytes(ite *IfdTagEntry) (value []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	value, err = ite.ValueBytes(ifd.addressableData, ifd.ByteOrder)
	log.PanicIf(err)

	return value, nil
}

// FindTagWithId returns a list of tags (usually just zero or one) that match
// the given tag ID. This is efficient.
func (ifd *Ifd) FindTagWithId(tagId uint16) (results []*IfdTagEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	results, found := ifd.EntriesByTagId[tagId]
	if found != true {
		log.Panic(ErrTagNotFound)
	}

	return results, nil
}

// FindTagWithName returns a list of tags (usually just zero or one) that match
// the given tag name. This is not efficient (though the labor is trivial).
func (ifd *Ifd) FindTagWithName(tagName string) (results []*IfdTagEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ii := ifd.Identity()
	it, err := ifd.tagIndex.GetWithName(ii, tagName)
	if log.Is(err, ErrTagNotFound) == true {
		log.Panic(ErrTagNotStandard)
	} else if err != nil {
		log.Panic(err)
	}

	results = make([]*IfdTagEntry, 0)
	for _, ite := range ifd.Entries {
		if ite.TagId == it.Id {
			results = append(results, ite)
		}
	}

	if len(results) == 0 {
		log.Panic(ErrTagNotFound)
	}

	return results, nil
}

func (ifd Ifd) String() string {
	parentOffset := uint32(0)
	if ifd.ParentIfd != nil {
		parentOffset = ifd.ParentIfd.Offset
	}

	return fmt.Sprintf("Ifd<ID=(%d) PARENT-IFD=[%s] IFD=[%s] INDEX=(%d) COUNT=(%d) OFF=(0x%04x) CHILDREN=(%d) PARENT=(0x%04x) NEXT-IFD=(0x%04x)>", ifd.Id, ifd.Ii.ParentIfdName, ifd.Ii.IfdName, ifd.Index, len(ifd.Entries), ifd.Offset, len(ifd.Children), parentOffset, ifd.NextIfdOffset)
}

func (ifd *Ifd) Thumbnail() (data []byte, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if ifd.thumbnailData == nil {
		log.Panic(ErrNoThumbnail)
	}

	return ifd.thumbnailData, nil
}

func (ifd *Ifd) Identity() IfdIdentity {
	return ifd.Ii
}

func (ifd *Ifd) dumpTags(tags []*IfdTagEntry) []*IfdTagEntry {
	if tags == nil {
		tags = make([]*IfdTagEntry, 0)
	}

	// Now, print the tags while also descending to child-IFDS as we encounter them.

	ifdsFoundCount := 0

	for _, tag := range ifd.Entries {
		tags = append(tags, tag)

		if tag.ChildIfdName != "" {
			ifdsFoundCount++

			childIfd, found := ifd.ChildIfdIndex[tag.ChildIfdName]
			if found != true {
				log.Panicf("alien child IFD referenced by a tag: [%s]", tag.ChildIfdName)
			}

			tags = childIfd.dumpTags(tags)
		}
	}

	if len(ifd.Children) != ifdsFoundCount {
		log.Panicf("have one or more dangling child IFDs: (%d) != (%d)", len(ifd.Children), ifdsFoundCount)
	}

	if ifd.NextIfd != nil {
		tags = ifd.NextIfd.dumpTags(tags)
	}

	return tags
}

// DumpTags prints the IFD hierarchy.
func (ifd *Ifd) DumpTags() []*IfdTagEntry {
	return ifd.dumpTags(nil)
}

func (ifd *Ifd) printTagTree(populateValues bool, index, level int, nextLink bool) {
	indent := strings.Repeat(" ", level*2)

	prefix := " "
	if nextLink {
		prefix = ">"
	}

	fmt.Printf("%s%sIFD: %s\n", indent, prefix, ifd)

	// Now, print the tags while also descending to child-IFDS as we encounter them.

	ifdsFoundCount := 0

	for _, tag := range ifd.Entries {
		if tag.ChildIfdName != "" {
			fmt.Printf("%s - TAG: %s\n", indent, tag)
		} else {
			it, err := ifd.tagIndex.Get(ifd.Identity(), tag.TagId)

			tagName := ""
			if err == nil {
				tagName = it.Name
			}

			var value interface{}
			if populateValues == true {
				var err error

				value, err = ifd.TagValue(tag)
				if err != nil {
					if log.Is(err, ErrUnhandledUnknownTypedTag) == true {
						value = UnparseableUnknownTagValuePlaceholder
					} else {
						log.Panic(err)
					}
				}
			}

			fmt.Printf("%s - TAG: %s NAME=[%s] VALUE=[%v]\n", indent, tag, tagName, value)
		}

		if tag.ChildIfdName != "" {
			ifdsFoundCount++

			childIfd, found := ifd.ChildIfdIndex[tag.ChildIfdName]
			if found != true {
				log.Panicf("alien child IFD referenced by a tag: [%s]", tag.ChildIfdName)
			}

			childIfd.printTagTree(populateValues, 0, level+1, false)
		}
	}

	if len(ifd.Children) != ifdsFoundCount {
		log.Panicf("have one or more dangling child IFDs: (%d) != (%d)", len(ifd.Children), ifdsFoundCount)
	}

	if ifd.NextIfd != nil {
		ifd.NextIfd.printTagTree(populateValues, index+1, level, true)
	}
}

// PrintTagTree prints the IFD hierarchy.
func (ifd *Ifd) PrintTagTree(populateValues bool) {
	ifd.printTagTree(populateValues, 0, 0, false)
}

func (ifd *Ifd) printIfdTree(level int, nextLink bool) {
	indent := strings.Repeat(" ", level*2)

	prefix := " "
	if nextLink {
		prefix = ">"
	}

	fmt.Printf("%s%s%s\n", indent, prefix, ifd)

	// Now, print the tags while also descending to child-IFDS as we encounter them.

	ifdsFoundCount := 0

	for _, tag := range ifd.Entries {
		if tag.ChildIfdName != "" {
			ifdsFoundCount++

			childIfd, found := ifd.ChildIfdIndex[tag.ChildIfdName]
			if found != true {
				log.Panicf("alien child IFD referenced by a tag: [%s]", tag.ChildIfdName)
			}

			childIfd.printIfdTree(level+1, false)
		}
	}

	if len(ifd.Children) != ifdsFoundCount {
		log.Panicf("have one or more dangling child IFDs: (%d) != (%d)", len(ifd.Children), ifdsFoundCount)
	}

	if ifd.NextIfd != nil {
		ifd.NextIfd.printIfdTree(level, true)
	}
}

// PrintIfdTree prints the IFD hierarchy.
func (ifd *Ifd) PrintIfdTree() {
	ifd.printIfdTree(0, false)
}

func (ifd *Ifd) dumpTree(tagsDump []string, level int) []string {
	if tagsDump == nil {
		tagsDump = make([]string, 0)
	}

	indent := strings.Repeat(" ", level*2)

	var ifdPhrase string
	if ifd.ParentIfd != nil {
		ifdPhrase = fmt.Sprintf("[%s]->[%s]:(%d)", ifd.ParentIfd.Ii.IfdName, ifd.Ii.IfdName, ifd.Index)
	} else {
		ifdPhrase = fmt.Sprintf("[ROOT]->[%s]:(%d)", ifd.Ii.IfdName, ifd.Index)
	}

	startBlurb := fmt.Sprintf("%s> IFD %s TOP", indent, ifdPhrase)
	tagsDump = append(tagsDump, startBlurb)

	ifdsFoundCount := 0
	for _, tag := range ifd.Entries {
		tagsDump = append(tagsDump, fmt.Sprintf("%s  - (0x%04x)", indent, tag.TagId))

		if tag.ChildIfdName != "" {
			ifdsFoundCount++

			childIfd, found := ifd.ChildIfdIndex[tag.ChildIfdName]
			if found != true {
				log.Panicf("alien child IFD referenced by a tag: [%s]", tag.ChildIfdName)
			}

			tagsDump = childIfd.dumpTree(tagsDump, level+1)
		}
	}

	if len(ifd.Children) != ifdsFoundCount {
		log.Panicf("have one or more dangling child IFDs: (%d) != (%d)", len(ifd.Children), ifdsFoundCount)
	}

	finishBlurb := fmt.Sprintf("%s< IFD %s BOTTOM", indent, ifdPhrase)
	tagsDump = append(tagsDump, finishBlurb)

	if ifd.NextIfd != nil {
		siblingBlurb := fmt.Sprintf("%s* LINKING TO SIBLING IFD [%s]:(%d)", indent, ifd.NextIfd.Ii.IfdName, ifd.NextIfd.Index)
		tagsDump = append(tagsDump, siblingBlurb)

		tagsDump = ifd.NextIfd.dumpTree(tagsDump, level)
	}

	return tagsDump
}

// DumpTree returns a list of strings describing the IFD hierarchy.
func (ifd *Ifd) DumpTree() []string {
	return ifd.dumpTree(nil, 0)
}

// GpsInfo parses and consolidates the GPS info. This can only be called on the
// GPS IFD.
func (ifd *Ifd) GpsInfo() (gi *GpsInfo, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Also add functionality to update the GPS info.

	gi = new(GpsInfo)

	if ifd.Ii != GpsIi {
		log.Panicf("GPS can only be read on GPS IFD: [%s] != [%s]", ifd.Ii, GpsIi)
	}

	if tags, found := ifd.EntriesByTagId[TagVersionId]; found == false {
		log.Panic(ErrNoGpsTags)
	} else if bytes.Compare(tags[0].value, []byte{2, 2, 0, 0}) != 0 {
		log.Panic(ErrNoGpsTags)
	}

	tags, found := ifd.EntriesByTagId[TagLatitudeId]
	if found == false {
		log.Panicf("latitude not found")
	}

	latitudeValue, err := ifd.TagValue(tags[0])
	log.PanicIf(err)

	tags, found = ifd.EntriesByTagId[TagLatitudeRefId]
	if found == false {
		log.Panicf("latitude-ref not found")
	}

	latitudeRefValue, err := ifd.TagValue(tags[0])
	log.PanicIf(err)

	tags, found = ifd.EntriesByTagId[TagLongitudeId]
	if found == false {
		log.Panicf("longitude not found")
	}

	longitudeValue, err := ifd.TagValue(tags[0])
	log.PanicIf(err)

	tags, found = ifd.EntriesByTagId[TagLongitudeRefId]
	if found == false {
		log.Panicf("longitude-ref not found")
	}

	longitudeRefValue, err := ifd.TagValue(tags[0])
	log.PanicIf(err)

	// Parse location.

	latitudeRaw := latitudeValue.([]Rational)

	gi.Latitude = GpsDegrees{
		Orientation: latitudeRefValue.(string)[0],
		Degrees:     int(float64(latitudeRaw[0].Numerator) / float64(latitudeRaw[0].Denominator)),
		Minutes:     int(float64(latitudeRaw[1].Numerator) / float64(latitudeRaw[1].Denominator)),
		Seconds:     int(float64(latitudeRaw[2].Numerator) / float64(latitudeRaw[2].Denominator)),
	}

	longitudeRaw := longitudeValue.([]Rational)

	gi.Longitude = GpsDegrees{
		Orientation: longitudeRefValue.(string)[0],
		Degrees:     int(float64(longitudeRaw[0].Numerator) / float64(longitudeRaw[0].Denominator)),
		Minutes:     int(float64(longitudeRaw[1].Numerator) / float64(longitudeRaw[1].Denominator)),
		Seconds:     int(float64(longitudeRaw[2].Numerator) / float64(longitudeRaw[2].Denominator)),
	}

	// Parse altitude.

	altitudeTags, foundAltitude := ifd.EntriesByTagId[TagAltitudeId]
	altitudeRefTags, foundAltitudeRef := ifd.EntriesByTagId[TagAltitudeRefId]

	if foundAltitude == true && foundAltitudeRef == true {
		altitudeValue, err := ifd.TagValue(altitudeTags[0])
		log.PanicIf(err)

		altitudeRefValue, err := ifd.TagValue(altitudeRefTags[0])
		log.PanicIf(err)

		altitudeRaw := altitudeValue.([]Rational)
		altitude := int(altitudeRaw[0].Numerator / altitudeRaw[0].Denominator)
		if altitudeRefValue.([]byte)[0] == 1 {
			altitude *= -1
		}

		gi.Altitude = altitude
	}

	// Parse time.

	timestampTags, foundTimestamp := ifd.EntriesByTagId[TagTimestampId]
	datestampTags, foundDatestamp := ifd.EntriesByTagId[TagDatestampId]

	if foundTimestamp == true && foundDatestamp == true {
		datestampValue, err := ifd.TagValue(datestampTags[0])
		log.PanicIf(err)

		dateParts := strings.Split(datestampValue.(string), ":")

		year, err1 := strconv.ParseUint(dateParts[0], 10, 16)
		month, err2 := strconv.ParseUint(dateParts[1], 10, 8)
		day, err3 := strconv.ParseUint(dateParts[2], 10, 8)

		if err1 == nil && err2 == nil && err3 == nil {
			timestampValue, err := ifd.TagValue(timestampTags[0])
			log.PanicIf(err)

			timestampRaw := timestampValue.([]Rational)

			hour := int(timestampRaw[0].Numerator / timestampRaw[0].Denominator)
			minute := int(timestampRaw[1].Numerator / timestampRaw[1].Denominator)
			second := int(timestampRaw[2].Numerator / timestampRaw[2].Denominator)

			gi.Timestamp = time.Date(int(year), time.Month(month), int(day), hour, minute, second, 0, time.UTC)
		}
	}

	return gi, nil
}

type ParsedTagVisitor func(*Ifd, *IfdTagEntry) error

func (ifd *Ifd) EnumerateTagsRecursively(visitor ParsedTagVisitor) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for ptr := ifd; ptr != nil; ptr = ptr.NextIfd {
		for _, ite := range ifd.Entries {
			if ite.ChildIfdName != "" {
				childIfd := ifd.ChildIfdIndex[ite.ChildIfdName]

				err := childIfd.EnumerateTagsRecursively(visitor)
				log.PanicIf(err)
			} else {
				err := visitor(ifd, ite)
				log.PanicIf(err)
			}
		}
	}

	return nil
}

type QueuedIfd struct {
	Ii    IfdIdentity
	TagId uint16

	Index  int
	Offset uint32
	Parent *Ifd

	// ParentTagIndex is our tag position in the parent IFD, if we had a parent
	// (if `ParentIfd` is not nil and we weren't an IFD referenced as a sibling
	// instead of as a child).
	ParentTagIndex int
}

type IfdIndex struct {
	RootIfd *Ifd
	Ifds    []*Ifd
	Tree    map[int]*Ifd
	Lookup  map[IfdIdentity][]*Ifd
}

// Scan enumerates the different EXIF blocks (called IFDs).
func (ie *IfdEnumerate) Collect(rootIfdOffset uint32, resolveValues bool) (index IfdIndex, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	tree := make(map[int]*Ifd)
	ifds := make([]*Ifd, 0)
	lookup := make(map[IfdIdentity][]*Ifd)

	queue := []QueuedIfd{
		{
			Ii:    RootIi,
			TagId: 0xffff,

			Index:  0,
			Offset: rootIfdOffset,
		},
	}

	edges := make(map[uint32]*Ifd)

	for {
		if len(queue) == 0 {
			break
		}

		qi := queue[0]

		ii := qi.Ii
		name := ii.IfdName

		index := qi.Index
		offset := qi.Offset
		parentIfd := qi.Parent

		queue = queue[1:]

		ifdEnumerateLogger.Debugf(nil, "Parsing IFD [%s] (%d) at offset (%04x).", ii.IfdName, index, offset)
		ite := ie.getTagEnumerator(offset)

		nextIfdOffset, entries, thumbnailData, err := ie.ParseIfd(ii, index, ite, nil, false, resolveValues)
		log.PanicIf(err)

		id := len(ifds)

		entriesByTagId := make(map[uint16][]*IfdTagEntry)
		for _, tag := range entries {
			tags, found := entriesByTagId[tag.TagId]
			if found == false {
				tags = make([]*IfdTagEntry, 0)
			}

			entriesByTagId[tag.TagId] = append(tags, tag)
		}

		ifd := &Ifd{
			addressableData: ie.exifData[ExifAddressableAreaStart:],

			ByteOrder: ie.byteOrder,

			Ii:    ii,
			TagId: qi.TagId,

			Id: id,

			ParentIfd:      parentIfd,
			ParentTagIndex: qi.ParentTagIndex,

			Name:           name,
			Index:          index,
			Offset:         offset,
			Entries:        entries,
			EntriesByTagId: entriesByTagId,

			// This is populated as each child is processed.
			Children: make([]*Ifd, 0),

			NextIfdOffset: nextIfdOffset,
			thumbnailData: thumbnailData,

			tagIndex: ie.tagIndex,
		}

		// Add ourselves to a big list of IFDs.
		ifds = append(ifds, ifd)

		// Install ourselves into a by-id lookup table (keys are unique).
		tree[id] = ifd

		// Install into by-name buckets.

		if list_, found := lookup[ii]; found == true {
			lookup[ii] = append(list_, ifd)
		} else {
			list_ = make([]*Ifd, 1)
			list_[0] = ifd

			lookup[ii] = list_
		}

		// Add a link from the previous IFD in the chain to us.
		if previousIfd, found := edges[offset]; found == true {
			previousIfd.NextIfd = ifd
		}

		// Attach as a child to our parent (where we appeared as a tag in
		// that IFD).
		if parentIfd != nil {
			parentIfd.Children = append(parentIfd.Children, ifd)
		}

		// Determine if any of our entries is a child IFD and queue it.
		for i, entry := range entries {
			if entry.ChildIfdName == "" {
				continue
			}

			childIi := IfdIdentity{
				ParentIfdName: name,
				IfdName:       entry.ChildIfdName,
			}

			qi := QueuedIfd{
				Ii:    childIi,
				TagId: entry.TagId,

				Index:          0,
				Offset:         entry.ValueOffset,
				Parent:         ifd,
				ParentTagIndex: i,
			}

			queue = append(queue, qi)
		}

		// If there's another IFD in the chain.
		if nextIfdOffset != 0 {
			// Allow the next link to know what the previous link was.
			edges[nextIfdOffset] = ifd

			qi := QueuedIfd{
				Ii:     ii,
				TagId:  0xffff,
				Index:  index + 1,
				Offset: nextIfdOffset,
			}

			queue = append(queue, qi)
		}
	}

	index.RootIfd = tree[0]
	index.Ifds = ifds
	index.Tree = tree
	index.Lookup = lookup

	err = ie.setChildrenIndex(index.RootIfd)
	log.PanicIf(err)

	return index, nil
}

func (ie *IfdEnumerate) setChildrenIndex(ifd *Ifd) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	childIfdIndex := make(map[string]*Ifd)
	for _, childIfd := range ifd.Children {
		childIfdIndex[childIfd.Ii.IfdName] = childIfd
	}

	ifd.ChildIfdIndex = childIfdIndex

	for _, childIfd := range ifd.Children {
		err := ie.setChildrenIndex(childIfd)
		log.PanicIf(err)
	}

	return nil
}

// ParseOneIfd is a hack to use an IE to parse a raw IFD block. Can be used for
// testing.
func ParseOneIfd(ii IfdIdentity, byteOrder binary.ByteOrder, ifdBlock []byte, visitor RawTagVisitor, resolveValues bool) (nextIfdOffset uint32, entries []*IfdTagEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ie := &IfdEnumerate{
		byteOrder: byteOrder,
	}

	ite := NewIfdTagEnumerator(ifdBlock, byteOrder, 0)

	nextIfdOffset, entries, _, err = ie.ParseIfd(ii, 0, ite, visitor, true, resolveValues)
	log.PanicIf(err)

	return nextIfdOffset, entries, nil
}

// ParseOneTag is a hack to use an IE to parse a raw tag block.
func ParseOneTag(ii IfdIdentity, byteOrder binary.ByteOrder, tagBlock []byte, resolveValue bool) (tag *IfdTagEntry, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ie := &IfdEnumerate{
		byteOrder: byteOrder,
	}

	ite := NewIfdTagEnumerator(tagBlock, byteOrder, 0)

	tag, err = ie.parseTag(ii, 0, ite, resolveValue)
	log.PanicIf(err)

	return tag, nil
}

func FindIfdFromRootIfd(rootIfd *Ifd, ifdDesignation string) (ifd *Ifd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ifd = rootIfd

	// TODO(dustin): !! Add test.

	switch ifdDesignation {
	case "ifd0":
		// We're already on it.

		return ifd, nil

	case "ifd1":
		if ifd.NextIfd == nil {
			log.Panicf("IFD1 not found")
		}

		return ifd.NextIfd, nil

	case "exif":
		ifd, err = ifd.ChildWithIfdIdentity(ExifIi)
		log.PanicIf(err)

		return ifd, nil

	case "iop":
		exifIfd, err := ifd.ChildWithIfdIdentity(ExifIi)
		log.PanicIf(err)

		ifd, err = exifIfd.ChildWithIfdIdentity(ExifIopIi)
		log.PanicIf(err)

		return ifd, nil

	case "gps":
		ifd, err = ifd.ChildWithIfdIdentity(GpsIi)
		log.PanicIf(err)

		return ifd, nil
	}

	candidates := make([]string, len(IfdDesignations))
	i := 0
	for key, _ := range IfdDesignations {
		candidates[i] = key
		i++
	}

	log.Panicf("IFD designation [%s] not valid. Use: %s\n", ifdDesignation, strings.Join(candidates, ", "))
	return nil, nil
}
