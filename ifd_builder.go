package exif

// NOTES:
//
// The thumbnail offset and length tags shouldn't be set directly. Use the
// (*IfdBuilder).SetThumbnail() method instead.

import (
	"errors"
	"fmt"
	"strings"

	"encoding/binary"

	log "github.com/dsoprea/go-logging"
)

var (
	ErrTagEntryNotFound = errors.New("tag entry not found")
	ErrChildIbNotFound  = errors.New("child IB not found")
)

type IfdBuilderTagValue struct {
	valueBytes []byte
	ib         *IfdBuilder
}

func (ibtv IfdBuilderTagValue) String() string {
	if ibtv.IsBytes() {
		var valuePhrase string
		if len(ibtv.valueBytes) <= 8 {
			valuePhrase = fmt.Sprintf("%v", ibtv.valueBytes)
		} else {
			valuePhrase = fmt.Sprintf("%v...", ibtv.valueBytes[:8])
		}

		return fmt.Sprintf("IfdBuilderTagValue<BYTES=%v LEN=(%d)>", valuePhrase, len(ibtv.valueBytes))
	} else if ibtv.IsIb() {
		return fmt.Sprintf("IfdBuilderTagValue<IB=%s>", ibtv.ib)
	} else {
		log.Panicf("IBTV state undefined")
		return ""
	}
}

func NewIfdBuilderTagValueFromBytes(valueBytes []byte) *IfdBuilderTagValue {
	return &IfdBuilderTagValue{
		valueBytes: valueBytes,
	}
}

func NewIfdBuilderTagValueFromIfdBuilder(ib *IfdBuilder) *IfdBuilderTagValue {
	return &IfdBuilderTagValue{
		ib: ib,
	}
}

// IsBytes returns true if the bytes are populated. This is always the case
// when we're loaded from a tag in an existing IFD.
func (ibtv IfdBuilderTagValue) IsBytes() bool {
	return ibtv.valueBytes != nil
}

func (ibtv IfdBuilderTagValue) Bytes() []byte {
	if !ibtv.IsBytes() {
		log.Panicf("this tag is not a byte-slice value")
	} else if ibtv.IsIb() {
		log.Panicf("this tag is an IFD-builder value not a byte-slice")
	}

	return ibtv.valueBytes
}

func (ibtv IfdBuilderTagValue) IsIb() bool {
	return ibtv.ib != nil
}

func (ibtv IfdBuilderTagValue) Ib() *IfdBuilder {
	if !ibtv.IsIb() {
		log.Panicf("this tag is not an IFD-builder value")
	} else if ibtv.IsBytes() {
		log.Panicf("this tag is a byte-slice, not a IFD-builder")
	}

	return ibtv.ib
}

type BuilderTag struct {
	// ifdPath is the path of the IFD that hosts this tag.
	ifdPath string

	tagId  uint16
	typeId TagTypePrimitive

	// value is either a value that can be encoded, an IfdBuilder instance (for
	// child IFDs), or an IfdTagEntry instance representing an existing,
	// previously-stored tag.
	value *IfdBuilderTagValue

	// byteOrder is the byte order. It's chiefly/originally here to support
	// printing the value.
	byteOrder binary.ByteOrder
}

func NewBuilderTag(ifdPath string, tagId uint16, typeId TagTypePrimitive, value *IfdBuilderTagValue, byteOrder binary.ByteOrder) *BuilderTag {
	return &BuilderTag{
		ifdPath:   ifdPath,
		tagId:     tagId,
		typeId:    typeId,
		value:     value,
		byteOrder: byteOrder,
	}
}

func NewChildIfdBuilderTag(ifdPath string, tagId uint16, value *IfdBuilderTagValue) *BuilderTag {
	return &BuilderTag{
		ifdPath: ifdPath,
		tagId:   tagId,
		typeId:  TypeLong,
		value:   value,
	}
}

func (bt *BuilderTag) Value() (value *IfdBuilderTagValue) {
	return bt.value
}

func (bt *BuilderTag) String() string {
	var valueString string

	if bt.value.IsBytes() {
		var err error

		valueString, err = Format(bt.value.Bytes(), bt.typeId, false, bt.byteOrder)
		log.PanicIf(err)
	} else {
		valueString = fmt.Sprintf("%v", bt.value)
	}

	return fmt.Sprintf("BuilderTag<IFD-PATH=[%s] TAG-ID=(0x%04x) TAG-TYPE=[%s] VALUE=[%s]>", bt.ifdPath, bt.tagId, TypeNames[bt.typeId], valueString)
}

func (bt *BuilderTag) SetValue(byteOrder binary.ByteOrder, value interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Add test.

	tt := NewTagType(bt.typeId, byteOrder)
	ve := NewValueEncoder(byteOrder)

	var ed EncodedData
	if bt.typeId == TypeUndefined {
		var err error

		ed, err = EncodeUndefined(bt.ifdPath, bt.tagId, value)
		log.PanicIf(err)
	} else {
		var err error

		ed, err = ve.EncodeWithType(tt, value)
		log.PanicIf(err)
	}

	bt.value = NewIfdBuilderTagValueFromBytes(ed.Encoded)

	return nil
}

// NewStandardBuilderTag constructs a `BuilderTag` instance. The type is looked
// up. `ii` is the type of IFD that owns this tag.
func NewStandardBuilderTag(ifdPath string, it *IndexedTag, byteOrder binary.ByteOrder, value interface{}) *BuilderTag {
	typeId := it.Type
	tt := NewTagType(typeId, byteOrder)

	ve := NewValueEncoder(byteOrder)

	var ed EncodedData
	if it.Type == TypeUndefined {
		var err error

		ed, err = EncodeUndefined(ifdPath, it.Id, value)
		log.PanicIf(err)
	} else {
		var err error

		ed, err = ve.EncodeWithType(tt, value)
		log.PanicIf(err)
	}

	tagValue := NewIfdBuilderTagValueFromBytes(ed.Encoded)

	return NewBuilderTag(
		ifdPath,
		it.Id,
		typeId,
		tagValue,
		byteOrder)
}

type IfdBuilder struct {
	// ifdName is the name of the IFD represented by this instance.
	name string

	// ifdPath is the path of the IFD represented by this instance.
	ifdPath string

	// fqIfdPath is the fully-qualified path of the IFD represented by this
	// instance.
	fqIfdPath string

	// ifdTagId will be non-zero if we're a child IFD.
	ifdTagId uint16

	byteOrder binary.ByteOrder

	// Includes both normal tags and IFD tags (which point to child IFDs).
	// TODO(dustin): Keep a separate list of children like with `Ifd`.
	// TODO(dustin): Either rename this or `Entries` in `Ifd` to be the same thing.
	tags []*BuilderTag

	// existingOffset will be the offset that this IFD is currently found at if
	// it represents an IFD that has previously been stored (or 0 if not).
	existingOffset uint32

	// nextIb represents the next link if we're chaining to another.
	nextIb *IfdBuilder

	// thumbnailData is populated with thumbnail data if there was thumbnail
	// data. Otherwise, it's nil.
	thumbnailData []byte

	ifdMapping *IfdMapping
	tagIndex   *TagIndex
}

func NewIfdBuilder(ifdMapping *IfdMapping, tagIndex *TagIndex, fqIfdPath string, byteOrder binary.ByteOrder) (ib *IfdBuilder) {
	ifdPath, err := ifdMapping.StripPathPhraseIndices(fqIfdPath)
	log.PanicIf(err)

	var ifdTagId uint16

	mi, err := ifdMapping.GetWithPath(ifdPath)
	if err == nil {
		ifdTagId = mi.TagId
	} else if !log.Is(err, ErrChildIfdNotMapped) {
		log.Panic(err)
	}

	ib = &IfdBuilder{
		// The right-most part of the IFD-path.
		name: mi.Name,

		// ifdPath describes the current IFD placement within the IFD tree.
		ifdPath: ifdPath,

		// fqIfdPath describes the current IFD placement within the IFD tree as
		// well as being qualified with non-zero indices.
		fqIfdPath: fqIfdPath,

		// ifdTagId is empty unless it's a child-IFD.
		ifdTagId: ifdTagId,

		byteOrder: byteOrder,
		tags:      make([]*BuilderTag, 0),

		ifdMapping: ifdMapping,
		tagIndex:   tagIndex,
	}

	return ib
}

// NewIfdBuilderWithExistingIfd creates a new IB using the same header type
// information as the given IFD.
func NewIfdBuilderWithExistingIfd(ifd *Ifd) (ib *IfdBuilder) {
	name := ifd.Name
	ifdPath := ifd.IfdPath
	fqIfdPath := ifd.FqIfdPath

	var ifdTagId uint16

	// There is no tag-ID for the root IFD. It will never be a child IFD.
	if ifdPath != IfdPathStandard {
		mi, err := ifd.ifdMapping.GetWithPath(ifdPath)
		log.PanicIf(err)

		ifdTagId = mi.TagId
	}

	ib = &IfdBuilder{
		name:           name,
		ifdPath:        ifdPath,
		fqIfdPath:      fqIfdPath,
		ifdTagId:       ifdTagId,
		byteOrder:      ifd.ByteOrder,
		existingOffset: ifd.Offset,
		ifdMapping:     ifd.ifdMapping,
		tagIndex:       ifd.tagIndex,
	}

	return ib
}

// NewIfdBuilderFromExistingChain creates a chain of IB instances from an
// IFD chain generated from real data.
func NewIfdBuilderFromExistingChain(rootIfd *Ifd, itevr *IfdTagEntryValueResolver) (firstIb *IfdBuilder) {
	// OBSOLETE(dustin): Support for `itevr` is now obsolete. This parameter will be removed in the future.

	var lastIb *IfdBuilder
	i := 0
	for thisExistingIfd := rootIfd; thisExistingIfd != nil; thisExistingIfd = thisExistingIfd.NextIfd {
		newIb := NewIfdBuilder(rootIfd.ifdMapping, rootIfd.tagIndex, rootIfd.FqIfdPath, thisExistingIfd.ByteOrder)
		if firstIb == nil {
			firstIb = newIb
		} else {
			lastIb.SetNextIb(newIb)
		}

		err := newIb.AddTagsFromExisting(thisExistingIfd, nil, nil, nil)
		log.PanicIf(err)

		lastIb = newIb
		i++
	}

	return firstIb
}

func (ib *IfdBuilder) NextIb() (nextIb *IfdBuilder, err error) {
	return ib.nextIb, nil
}

func (ib *IfdBuilder) ChildWithTagId(childIfdTagId uint16) (childIb *IfdBuilder, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for _, bt := range ib.tags {
		if !bt.value.IsIb() {
			continue
		}

		childIbThis := bt.value.Ib()

		if childIbThis.ifdTagId == childIfdTagId {
			return childIbThis, nil
		}
	}

	log.Panic(ErrChildIbNotFound)

	// Never reached.
	return nil, nil
}

func getOrCreateIbFromRootIbInner(rootIb *IfdBuilder, parentIb *IfdBuilder, currentLineage []IfdTagIdAndIndex) (ib *IfdBuilder, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Add test.

	thisIb := rootIb

	// Since we're calling ourselves recursively with incrementally different
	// paths, the FQ IFD-path of the parent that called us needs to be passed
	// in, in order for us to know it.
	var parentLineage []IfdTagIdAndIndex
	if parentIb != nil {
		var err error

		parentLineage, err = thisIb.ifdMapping.ResolvePath(parentIb.fqIfdPath)
		log.PanicIf(err)
	}

	// Process the current path part.
	currentItii := currentLineage[0]

	// Make sure the leftmost part of the FQ IFD-path agrees with the IB we
	// were given.

	expectedFqRootIfdPath := ""
	if parentLineage != nil {
		expectedLineage := append(parentLineage, currentItii)
		expectedFqRootIfdPath = thisIb.ifdMapping.PathPhraseFromLineage(expectedLineage)
	} else {
		expectedFqRootIfdPath = thisIb.ifdMapping.PathPhraseFromLineage(currentLineage[:1])
	}

	if expectedFqRootIfdPath != thisIb.fqIfdPath {
		log.Panicf("the FQ IFD-path [%s] we were given does not match the builder's FQ IFD-path [%s]", expectedFqRootIfdPath, thisIb.fqIfdPath)
	}

	// If we actually wanted a sibling (currentItii.Index > 0) then seek to it,
	// appending new siblings, as required, until we get there.
	for i := 0; i < currentItii.Index; i++ {
		if thisIb.nextIb == nil {
			// Generate an FQ IFD-path for the sibling. It'll use the same
			// non-FQ IFD-path as the current IB.

			siblingFqIfdPath := ""
			if parentLineage != nil {
				siblingFqIfdPath = fmt.Sprintf("%s/%s%d", parentIb.fqIfdPath, currentItii.Name, i+1)
			} else {
				siblingFqIfdPath = fmt.Sprintf("%s%d", currentItii.Name, i+1)
			}

			thisIb.nextIb = NewIfdBuilder(thisIb.ifdMapping, thisIb.tagIndex, siblingFqIfdPath, thisIb.byteOrder)
		}

		thisIb = thisIb.nextIb
	}

	// There is no child IFD to process. We're done.
	if len(currentLineage) == 1 {
		return thisIb, nil
	}

	// Establish the next child to be processed.

	childItii := currentLineage[1]

	var foundChild *IfdBuilder
	for _, bt := range thisIb.tags {
		if !bt.value.IsIb() {
			continue
		}

		childIb := bt.value.Ib()

		if childIb.ifdTagId == childItii.TagId {
			foundChild = childIb
			break
		}
	}

	// If we didn't find the child, add it.
	if foundChild == nil {
		thisIbLineage, err := thisIb.ifdMapping.ResolvePath(thisIb.fqIfdPath)
		log.PanicIf(err)

		childLineage := make([]IfdTagIdAndIndex, len(thisIbLineage)+1)
		copy(childLineage, thisIbLineage)

		childLineage[len(childLineage)-1] = childItii

		fqIfdChildPath := thisIb.ifdMapping.FqPathPhraseFromLineage(childLineage)

		foundChild = NewIfdBuilder(thisIb.ifdMapping, thisIb.tagIndex, fqIfdChildPath, thisIb.byteOrder)

		err = thisIb.AddChildIb(foundChild)
		log.PanicIf(err)
	}

	finalIb, err := getOrCreateIbFromRootIbInner(foundChild, thisIb, currentLineage[1:])
	log.PanicIf(err)

	return finalIb, nil
}

// GetOrCreateIbFromRootIb returns an IB representing the requested IFD, even if
// an IB doesn't already exist for it. This function may call itself
// recursively.
func GetOrCreateIbFromRootIb(rootIb *IfdBuilder, fqIfdPath string) (ib *IfdBuilder, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// lineage is a necessity of our recursion process. It doesn't include any
	// parent IFDs on its left-side; it starts with the current IB only.
	lineage, err := rootIb.ifdMapping.ResolvePath(fqIfdPath)
	log.PanicIf(err)

	ib, err = getOrCreateIbFromRootIbInner(rootIb, nil, lineage)
	log.PanicIf(err)

	return ib, nil
}

func (ib *IfdBuilder) String() string {
	nextIfdPhrase := ""
	if ib.nextIb != nil {
		// TODO(dustin): We were setting this to ii.String(), but we were getting hex-data when printing this after building from an existing chain.
		nextIfdPhrase = ib.nextIb.ifdPath
	}

	return fmt.Sprintf("IfdBuilder<PATH=[%s] TAG-ID=(0x%04x) COUNT=(%d) OFF=(0x%04x) NEXT-IFD-PATH=[%s]>", ib.ifdPath, ib.ifdTagId, len(ib.tags), ib.existingOffset, nextIfdPhrase)
}

func (ib *IfdBuilder) Tags() (tags []*BuilderTag) {
	return ib.tags
}

// SetThumbnail sets thumbnail data.
//
// NOTES:
//
// - We don't manage any facet of the thumbnail data. This is the
//   responsibility of the user/developer.
// - This method will fail unless the thumbnail is set on a the root IFD.
//   However, in order to be valid, it must be set on the second one, linked to
//   by the first, as per the EXIF/TIFF specification.
// - We set the offset to (0) now but will allocate the data and properly assign
//   the offset when the IB is encoded (later).
func (ib *IfdBuilder) SetThumbnail(data []byte) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if ib.ifdPath != IfdPathStandard {
		log.Panicf("thumbnails can only go into a root Ifd (and only the second one)")
	}

	// TODO(dustin): !! Add a test for this function.

	if len(data) == 0 {
		log.Panic("thumbnail is empty")
	}

	ib.thumbnailData = data

	ibtvfb := NewIfdBuilderTagValueFromBytes(ib.thumbnailData)
	offsetBt :=
		NewBuilderTag(
			ib.ifdPath,
			ThumbnailOffsetTagId,
			TypeLong,
			ibtvfb,
			ib.byteOrder)

	err = ib.Set(offsetBt)
	log.PanicIf(err)

	thumbnailSizeIt, err := ib.tagIndex.Get(ib.ifdPath, ThumbnailSizeTagId)
	log.PanicIf(err)

	sizeBt := NewStandardBuilderTag(ib.ifdPath, thumbnailSizeIt, ib.byteOrder, []uint32{uint32(len(ib.thumbnailData))})

	err = ib.Set(sizeBt)
	log.PanicIf(err)

	return nil
}

func (ib *IfdBuilder) Thumbnail() []byte {
	return ib.thumbnailData
}

func (ib *IfdBuilder) printTagTree(levels int) {
	indent := strings.Repeat(" ", levels*2)

	i := 0
	for currentIb := ib; currentIb != nil; currentIb = currentIb.nextIb {
		prefix := " "
		if i > 0 {
			prefix = ">"
		}

		if levels == 0 {
			fmt.Printf("%s%sIFD: %s INDEX=(%d)\n", indent, prefix, currentIb, i)
		} else {
			fmt.Printf("%s%sChild IFD: %s\n", indent, prefix, currentIb)
		}

		if len(currentIb.tags) > 0 {
			fmt.Printf("\n")

			for i, tag := range currentIb.tags {
				isChildIb := false
				_, err := ib.ifdMapping.GetChild(currentIb.ifdPath, tag.tagId)
				if err == nil {
					isChildIb = true
				} else if !log.Is(err, ErrChildIfdNotMapped) {
					log.Panic(err)
				}

				tagName := ""

				// If a normal tag (not a child IFD) get the name.
				if isChildIb {
					tagName = "<Child IFD>"
				} else {
					it, err := ib.tagIndex.Get(tag.ifdPath, tag.tagId)
					if log.Is(err, ErrTagNotFound) {
						tagName = "<UNKNOWN>"
					} else if err != nil {
						log.Panic(err)
					} else {
						tagName = it.Name
					}
				}

				value := tag.Value()

				if value.IsIb() {
					fmt.Printf("%s  (%d): [%s] %s\n", indent, i, tagName, value.Ib())
				} else {
					fmt.Printf("%s  (%d): [%s] %s\n", indent, i, tagName, tag)
				}

				if isChildIb {
					if !tag.value.IsIb() {
						log.Panicf("tag-ID (0x%04x) is an IFD but the tag value is not an IB instance: %v", tag.tagId, tag)
					}

					fmt.Printf("\n")

					childIb := tag.value.Ib()
					childIb.printTagTree(levels + 1)
				}
			}

			fmt.Printf("\n")
		}

		i++
	}
}

func (ib *IfdBuilder) PrintTagTree() {
	ib.printTagTree(0)
}

func (ib *IfdBuilder) printIfdTree(levels int) {
	indent := strings.Repeat(" ", levels*2)

	i := 0
	for currentIb := ib; currentIb != nil; currentIb = currentIb.nextIb {
		prefix := " "
		if i > 0 {
			prefix = ">"
		}

		fmt.Printf("%s%s%s\n", indent, prefix, currentIb)

		if len(currentIb.tags) > 0 {
			for _, tag := range currentIb.tags {
				isChildIb := false
				_, err := ib.ifdMapping.GetChild(currentIb.ifdPath, tag.tagId)
				if err == nil {
					isChildIb = true
				} else if !log.Is(err, ErrChildIfdNotMapped) {
					log.Panic(err)
				}

				if isChildIb {
					if !tag.value.IsIb() {
						log.Panicf("tag-ID (0x%04x) is an IFD but the tag value is not an IB instance: %v", tag.tagId, tag)
					}

					childIb := tag.value.Ib()
					childIb.printIfdTree(levels + 1)
				}
			}
		}

		i++
	}
}

func (ib *IfdBuilder) PrintIfdTree() {
	ib.printIfdTree(0)
}

func (ib *IfdBuilder) dumpToStrings(thisIb *IfdBuilder, prefix string, tagId uint16, lines []string) (linesOutput []string) {
	if lines == nil {
		linesOutput = make([]string, 0)
	} else {
		linesOutput = lines
	}

	siblingIfdIndex := 0
	for ; thisIb != nil; thisIb = thisIb.nextIb {
		line := fmt.Sprintf("IFD<PARENTS=[%s] FQ-IFD-PATH=[%s] IFD-INDEX=(%d) IFD-TAG-ID=(0x%04x) TAG=[0x%04x]>", prefix, thisIb.fqIfdPath, siblingIfdIndex, thisIb.ifdTagId, tagId)
		linesOutput = append(linesOutput, line)

		for i, tag := range thisIb.tags {
			var childIb *IfdBuilder
			childIfdName := ""
			if tag.value.IsIb() {
				childIb = tag.value.Ib()
				childIfdName = childIb.ifdPath
			}

			line := fmt.Sprintf("TAG<PARENTS=[%s] FQ-IFD-PATH=[%s] IFD-TAG-ID=(0x%04x) CHILD-IFD=[%s] TAG-INDEX=(%d) TAG=[0x%04x]>", prefix, thisIb.fqIfdPath, thisIb.ifdTagId, childIfdName, i, tag.tagId)
			linesOutput = append(linesOutput, line)

			if childIb == nil {
				continue
			}

			childPrefix := ""
			if prefix == "" {
				childPrefix = thisIb.ifdPath
			} else {
				childPrefix = fmt.Sprintf("%s->%s", prefix, thisIb.ifdPath)
			}

			linesOutput = thisIb.dumpToStrings(childIb, childPrefix, tag.tagId, linesOutput)
		}

		siblingIfdIndex++
	}

	return linesOutput
}

func (ib *IfdBuilder) DumpToStrings() (lines []string) {
	return ib.dumpToStrings(ib, "", 0, lines)
}

func (ib *IfdBuilder) SetNextIb(nextIb *IfdBuilder) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ib.nextIb = nextIb

	return nil
}

func (ib *IfdBuilder) DeleteN(tagId uint16, n int) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if n < 1 {
		log.Panicf("N must be at least 1: (%d)", n)
	}

	for n > 0 {
		j := -1
		for i, bt := range ib.tags {
			if bt.tagId == tagId {
				j = i
				break
			}
		}

		if j == -1 {
			log.Panic(ErrTagEntryNotFound)
		}

		ib.tags = append(ib.tags[:j], ib.tags[j+1:]...)
		n--
	}

	return nil
}

func (ib *IfdBuilder) DeleteFirst(tagId uint16) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = ib.DeleteN(tagId, 1)
	log.PanicIf(err)

	return nil
}

func (ib *IfdBuilder) DeleteAll(tagId uint16) (n int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	for {
		err = ib.DeleteN(tagId, 1)
		if log.Is(err, ErrTagEntryNotFound) {
			break
		} else if err != nil {
			log.Panic(err)
		}

		n++
	}

	return n, nil
}

func (ib *IfdBuilder) ReplaceAt(position int, bt *BuilderTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if position < 0 {
		log.Panicf("replacement position must be 0 or greater")
	} else if position >= len(ib.tags) {
		log.Panicf("replacement position does not exist")
	}

	ib.tags[position] = bt

	return nil
}

func (ib *IfdBuilder) Replace(tagId uint16, bt *BuilderTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	position, err := ib.Find(tagId)
	log.PanicIf(err)

	ib.tags[position] = bt

	return nil
}

// Set will add a new entry or update an existing entry.
func (ib *IfdBuilder) Set(bt *BuilderTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	position, err := ib.Find(bt.tagId)
	if err == nil {
		ib.tags[position] = bt
	} else if log.Is(err, ErrTagEntryNotFound) {
		err = ib.add(bt)
		log.PanicIf(err)
	} else {
		log.Panic(err)
	}

	return nil
}

func (ib *IfdBuilder) FindN(tagId uint16, maxFound int) (found []int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	found = make([]int, 0)

	for i, bt := range ib.tags {
		if bt.tagId == tagId {
			found = append(found, i)
			if maxFound == 0 || len(found) >= maxFound {
				break
			}
		}
	}

	return found, nil
}

func (ib *IfdBuilder) Find(tagId uint16) (position int, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	found, err := ib.FindN(tagId, 1)
	log.PanicIf(err)

	if len(found) == 0 {
		log.Panic(ErrTagEntryNotFound)
	}

	return found[0], nil
}

func (ib *IfdBuilder) FindTag(tagId uint16) (bt *BuilderTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	found, err := ib.FindN(tagId, 1)
	log.PanicIf(err)

	if len(found) == 0 {
		log.Panic(ErrTagEntryNotFound)
	}

	position := found[0]

	return ib.tags[position], nil
}

func (ib *IfdBuilder) FindTagWithName(tagName string) (bt *BuilderTag, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	it, err := ib.tagIndex.GetWithName(ib.ifdPath, tagName)
	log.PanicIf(err)

	found, err := ib.FindN(it.Id, 1)
	log.PanicIf(err)

	if len(found) == 0 {
		log.Panic(ErrTagEntryNotFound)
	}

	position := found[0]

	return ib.tags[position], nil
}

func (ib *IfdBuilder) add(bt *BuilderTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if bt.ifdPath == "" {
		log.Panicf("BuilderTag ifdPath is not set: %s", bt)
	} else if bt.typeId == 0x0 {
		log.Panicf("BuilderTag type-ID is not set: %s", bt)
	} else if bt.value == nil {
		log.Panicf("BuilderTag value is not set: %s", bt)
	}

	ib.tags = append(ib.tags, bt)
	return nil
}

func (ib *IfdBuilder) Add(bt *BuilderTag) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if bt.value.IsIb() {
		log.Panicf("child IfdBuilders must be added via AddChildIb() or AddTagsFromExisting(), not Add()")
	}

	err = ib.add(bt)
	log.PanicIf(err)

	return nil
}

// AddChildIb adds a tag that branches to a new IFD.
func (ib *IfdBuilder) AddChildIb(childIb *IfdBuilder) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	if childIb.ifdTagId == 0 {
		log.Panicf("IFD can not be used as a child IFD (not associated with a tag-ID): %v", childIb)
	} else if childIb.byteOrder != ib.byteOrder {
		log.Panicf("Child IFD does not have the same byte-order: [%s] != [%s]", childIb.byteOrder, ib.byteOrder)
	}

	// Since no standard IFDs supports occuring more than once, check that a
	// tag of this type has not been previously added. Note that we just search
	// the current IFD and *not every* IFD.
	for _, bt := range childIb.tags {
		if bt.tagId == childIb.ifdTagId {
			log.Panicf("child-IFD already added: %v", childIb.ifdPath)
		}
	}

	bt := ib.NewBuilderTagFromBuilder(childIb)
	ib.tags = append(ib.tags, bt)

	return nil
}

func (ib *IfdBuilder) NewBuilderTagFromBuilder(childIb *IfdBuilder) (bt *BuilderTag) {
	defer func() {
		if state := recover(); state != nil {
			err := log.Wrap(state.(error))
			log.Panic(err)
		}
	}()

	value := NewIfdBuilderTagValueFromIfdBuilder(childIb)

	bt = NewChildIfdBuilderTag(
		ib.ifdPath,
		childIb.ifdTagId,
		value)

	return bt
}

// AddTagsFromExisting does a verbatim copy of the entries in `ifd` to this
// builder. It excludes child IFDs. These must be added explicitly via
// `AddChildIb()`.
func (ib *IfdBuilder) AddTagsFromExisting(ifd *Ifd, itevr *IfdTagEntryValueResolver, includeTagIds []uint16, excludeTagIds []uint16) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// OBSOLETE(dustin): Support for `itevr` is now obsolete. This parameter will be removed in the future.

	thumbnailData, err := ifd.Thumbnail()
	if err == nil {
		err = ib.SetThumbnail(thumbnailData)
		log.PanicIf(err)
	} else if !log.Is(err, ErrNoThumbnail) {
		log.Panic(err)
	}

	for i, ite := range ifd.Entries {
		if ite.TagId == ThumbnailOffsetTagId || ite.TagId == ThumbnailSizeTagId {
			// These will be added on-the-fly when we encode.
			continue
		}

		if len(excludeTagIds) > 0 {
			found := false
			for _, excludedTagId := range excludeTagIds {
				if excludedTagId == ite.TagId {
					found = true
				}
			}

			if found {
				continue
			}
		}

		if len(includeTagIds) > 0 {
			// Whether or not there was a list of excludes, if there is a list
			// of includes than the current tag has to be in it.

			found := false
			for _, includedTagId := range includeTagIds {
				if includedTagId == ite.TagId {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		var bt *BuilderTag

		if ite.ChildIfdPath != "" {
			// If we want to add an IFD tag, we'll have to build it first and
			// *then* add it via a different method.

			// Figure out which of the child-IFDs that are associated with
			// this IFD represents this specific child IFD.

			var childIfd *Ifd
			for _, thisChildIfd := range ifd.Children {
				if thisChildIfd.ParentTagIndex != i {
					continue
				} else if thisChildIfd.TagId != 0xffff && thisChildIfd.TagId != ite.TagId {
					log.Panicf("child-IFD tag is not correct: TAG-POSITION=(%d) ITE=%s CHILD-IFD=%s", thisChildIfd.ParentTagIndex, ite, thisChildIfd)
				}

				childIfd = thisChildIfd
				break
			}

			if childIfd == nil {
				childTagIds := make([]string, len(ifd.Children))
				for j, childIfd := range ifd.Children {
					childTagIds[j] = fmt.Sprintf("0x%04x (parent tag-position %d)", childIfd.TagId, childIfd.ParentTagIndex)
				}

				log.Panicf("could not find child IFD for child ITE: IFD-PATH=[%s] TAG-ID=(0x%04x) CURRENT-TAG-POSITION=(%d) CHILDREN=%v", ite.IfdPath, ite.TagId, i, childTagIds)
			}

			childIb := NewIfdBuilderFromExistingChain(childIfd, nil)
			bt = ib.NewBuilderTagFromBuilder(childIb)
		} else {
			// Non-IFD tag.

			valueContext := ifd.GetValueContext(ite)

			var rawBytes []byte

			if ite.TagType == TypeUndefined {
				// It's an undefined-type value. Try to process, or skip if
				// we don't know how to.

				undefinedInterface, err := valueContext.Undefined()
				if err != nil {
					if err == ErrUnhandledUnknownTypedTag {
						// It's an undefined-type tag that we don't handle. If
						// we don't know how to handle it, we can't know how
						// many bytes it is and we must skip it.
						continue
					}

					log.Panic(err)
				}

				undefined, ok := undefinedInterface.(UnknownTagValue)
				if !ok {
					log.Panicf("unexpected value returned from undefined-value processor")
				}

				rawBytes, err = undefined.ValueBytes()
				log.PanicIf(err)
			} else {
				// It's a value with a standard type.

				var err error

				rawBytes, err = valueContext.readRawEncoded()
				log.PanicIf(err)
			}

			value := NewIfdBuilderTagValueFromBytes(rawBytes)

			bt = NewBuilderTag(
				ifd.IfdPath,
				ite.TagId,
				ite.TagType,
				value,
				ib.byteOrder)
		}

		err := ib.add(bt)
		log.PanicIf(err)
	}

	return nil
}

// AddStandard quickly and easily composes and adds the tag using the
// information already known about a tag. Only works with standard tags.
func (ib *IfdBuilder) AddStandard(tagId uint16, value interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	it, err := ib.tagIndex.Get(ib.ifdPath, tagId)
	log.PanicIf(err)

	bt := NewStandardBuilderTag(ib.ifdPath, it, ib.byteOrder, value)

	err = ib.add(bt)
	log.PanicIf(err)

	return nil
}

// AddStandardWithName quickly and easily composes and adds the tag using the
// information already known about a tag (using the name). Only works with
// standard tags.
func (ib *IfdBuilder) AddStandardWithName(tagName string, value interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	it, err := ib.tagIndex.GetWithName(ib.ifdPath, tagName)
	log.PanicIf(err)

	bt := NewStandardBuilderTag(ib.ifdPath, it, ib.byteOrder, value)

	err = ib.add(bt)
	log.PanicIf(err)

	return nil
}

// SetStandard quickly and easily composes and adds or replaces the tag using
// the information already known about a tag. Only works with standard tags.
func (ib *IfdBuilder) SetStandard(tagId uint16, value interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Add test for this function.

	it, err := ib.tagIndex.Get(ib.ifdPath, tagId)
	log.PanicIf(err)

	bt := NewStandardBuilderTag(ib.ifdPath, it, ib.byteOrder, value)

	i, err := ib.Find(tagId)
	if err != nil {
		if !log.Is(err, ErrTagEntryNotFound) {
			log.Panic(err)
		}

		ib.tags = append(ib.tags, bt)
	} else {
		ib.tags[i] = bt
	}

	return nil
}

// SetStandardWithName quickly and easily composes and adds or replaces the
// tag using the information already known about a tag (using the name). Only
// works with standard tags.
func (ib *IfdBuilder) SetStandardWithName(tagName string, value interface{}) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	// TODO(dustin): !! Add test for this function.

	it, err := ib.tagIndex.GetWithName(ib.ifdPath, tagName)
	log.PanicIf(err)

	bt := NewStandardBuilderTag(ib.ifdPath, it, ib.byteOrder, value)

	i, err := ib.Find(bt.tagId)
	if err != nil {
		if !log.Is(err, ErrTagEntryNotFound) {
			log.Panic(err)
		}

		ib.tags = append(ib.tags, bt)
	} else {
		ib.tags[i] = bt
	}

	return nil
}
