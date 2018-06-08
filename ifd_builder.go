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

    "github.com/dsoprea/go-logging"
)

var (
    ifdBuilderLogger = log.NewLogger("exif.ifd_builder")
)

var (
    ErrTagEntryNotFound = errors.New("tag entry not found")
)

type IfdBuilderTagValue struct {
    valueBytes []byte
    ib         *IfdBuilder
}

func (ibtv IfdBuilderTagValue) String() string {
    if ibtv.IsBytes() == true {
        var valuePhrase string
        if len(ibtv.valueBytes) <= 8 {
            valuePhrase = fmt.Sprintf("%v", ibtv.valueBytes)
        } else {
            valuePhrase = fmt.Sprintf("%v...", ibtv.valueBytes[:8])
        }

        return fmt.Sprintf("IfdBuilderTagValue<BYTES=%v LEN=(%d)>", valuePhrase, len(ibtv.valueBytes))
    } else if ibtv.IsIb() == true {
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

func (ibtv IfdBuilderTagValue) IsBytes() bool {
    return ibtv.valueBytes != nil
}

func (ibtv IfdBuilderTagValue) Bytes() []byte {
    if ibtv.IsBytes() == false {
        log.Panicf("this tag is not a byte-slice value")
    }

    return ibtv.valueBytes
}

func (ibtv IfdBuilderTagValue) IsIb() bool {
    return ibtv.ib != nil
}

func (ibtv IfdBuilderTagValue) Ib() *IfdBuilder {
    if ibtv.IsIb() == false {
        log.Panicf("this tag is not an IFD-builder value")
    }

    return ibtv.ib
}

type BuilderTag struct {
    // ii is the IfdIdentity of the IFD that hosts this tag.
    ii IfdIdentity

    tagId  uint16
    typeId uint16

    // value is either a value that can be encoded, an IfdBuilder instance (for
    // child IFDs), or an IfdTagEntry instance representing an existing,
    // previously-stored tag.
    value *IfdBuilderTagValue
}

func NewBuilderTag(ii IfdIdentity, tagId uint16, typeId uint16, value *IfdBuilderTagValue) *BuilderTag {
    return &BuilderTag{
        ii:     ii,
        tagId:  tagId,
        typeId: typeId,
        value:  value,
    }
}

func NewChildIfdBuilderTag(ii IfdIdentity, tagId uint16, value *IfdBuilderTagValue) *BuilderTag {
    return &BuilderTag{
        ii:     ii,
        tagId:  tagId,
        typeId: TypeLong,
        value:  value,
    }
}

func (bt *BuilderTag) Value() (value *IfdBuilderTagValue) {
    return bt.value
}

func (bt *BuilderTag) String() string {
    return fmt.Sprintf("BuilderTag<IFD=[%s] TAG-ID=(0x%04x) TAG-TYPE=[%s] VALUE=[%s]>", bt.ii, bt.tagId, TypeNames[bt.typeId], bt.value)
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

        ed, err = EncodeUndefined(bt.ii, bt.tagId, value)
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
func NewStandardBuilderTag(ii IfdIdentity, tagId uint16, byteOrder binary.ByteOrder, value interface{}) *BuilderTag {
    it, err := tagIndex.Get(ii, tagId)
    log.PanicIf(err)

    typeId := it.Type
    tt := NewTagType(typeId, byteOrder)

    ve := NewValueEncoder(byteOrder)

    var ed EncodedData
    if it.Type == TypeUndefined {
        var err error

        ed, err = EncodeUndefined(ii, tagId, value)
        log.PanicIf(err)
    } else {
        var err error

        ed, err = ve.EncodeWithType(tt, value)
        log.PanicIf(err)
    }

    tagValue := NewIfdBuilderTagValueFromBytes(ed.Encoded)

    return NewBuilderTag(
        ii,
        tagId,
        typeId,
        tagValue)
}

// NewStandardBuilderTagWithName allows us to easily generate solid, consistent
// tags for testing with. `ii` is the type of IFD that owns this tag. This can
// not be an IFD (IFDs are not associated with standardized, official names).
func NewStandardBuilderTagWithName(ii IfdIdentity, tagName string, byteOrder binary.ByteOrder, value interface{}) *BuilderTag {
    it, err := tagIndex.GetWithName(ii, tagName)
    log.PanicIf(err)

    tt := NewTagType(it.Type, byteOrder)

    ve := NewValueEncoder(byteOrder)

    var ed EncodedData
    if it.Type == TypeUndefined {
        var err error

        ed, err = EncodeUndefined(ii, it.Id, value)
        log.PanicIf(err)
    } else {
        var err error

        ed, err = ve.EncodeWithType(tt, value)
        log.PanicIf(err)
    }

    tagValue := NewIfdBuilderTagValueFromBytes(ed.Encoded)

    return NewBuilderTag(
        ii,
        it.Id,
        it.Type,
        tagValue)
}

type IfdBuilder struct {
    // ifd is the IfdIdentity instance of the IFD that owns the current tag.
    ii IfdIdentity

    // ifdTagId will be non-zero if we're a child IFD.
    ifdTagId uint16

    byteOrder binary.ByteOrder

    // Includes both normal tags and IFD tags (which point to child IFDs).
    tags []*BuilderTag

    // existingOffset will be the offset that this IFD is currently found at if
    // it represents an IFD that has previously been stored (or 0 if not).
    existingOffset uint32

    // nextIb represents the next link if we're chaining to another.
    nextIb *IfdBuilder

    // thumbnailData is populated with thumbnail data if there was thumbnail
    // data. Otherwise, it's nil.
    thumbnailData []byte
}

func NewIfdBuilder(ii IfdIdentity, byteOrder binary.ByteOrder) (ib *IfdBuilder) {
    ifdTagId, _ := IfdTagIdWithIdentity(ii)

    ib = &IfdBuilder{
        // ii describes the current IFD and its parent.
        ii: ii,

        // ifdTagId is empty unless it's a child-IFD.
        ifdTagId: ifdTagId,

        byteOrder: byteOrder,
        tags:      make([]*BuilderTag, 0),
    }

    return ib
}

// NewIfdBuilderWithExistingIfd creates a new IB using the same header type
// information as the given IFD.
func NewIfdBuilderWithExistingIfd(ifd *Ifd) (ib *IfdBuilder) {
    ii := ifd.Identity()

    var ifdTagId uint16

    // There is no tag-ID for the root IFD. It will never be a child IFD.
    if ii != RootIi {
        ifdTagId = IfdTagIdWithIdentityOrFail(ii)
    }

    ib = &IfdBuilder{
        ii:             ii,
        ifdTagId:       ifdTagId,
        byteOrder:      ifd.ByteOrder,
        existingOffset: ifd.Offset,
    }

    return ib
}

// NewIfdBuilderFromExistingChain creates a chain of IB instances from an
// IFD chain generated from real data.
func NewIfdBuilderFromExistingChain(rootIfd *Ifd, itevr *IfdTagEntryValueResolver) (firstIb *IfdBuilder) {
    // TODO(dustin): !! When we actually write the code to flatten the IB to bytes, make sure to skip the tags that have a nil value (which will happen when we add-from-exsting without a resolver instance).

    var lastIb *IfdBuilder
    i := 0
    for thisExistingIfd := rootIfd; thisExistingIfd != nil; thisExistingIfd = thisExistingIfd.NextIfd {
        ii := thisExistingIfd.Identity()

        newIb := NewIfdBuilder(ii, thisExistingIfd.ByteOrder)
        if firstIb == nil {
            firstIb = newIb
        } else {
            lastIb.SetNextIb(newIb)
        }

        err := newIb.AddTagsFromExisting(thisExistingIfd, itevr, nil, nil)
        log.PanicIf(err)

        lastIb = newIb
        i++
    }

    return firstIb
}

func (ib *IfdBuilder) String() string {
    nextIfdPhrase := ""
    if ib.nextIb != nil {
        // TODO(dustin): We were setting this to ii.String(), but we were getting hex-data when printing this after building from an existing chain.
        nextIfdPhrase = ib.nextIb.ii.IfdName
    }

    return fmt.Sprintf("IfdBuilder<PARENT-IFD=[%s] IFD=[%s] TAG-ID=(0x%04x) COUNT=(%d) OFF=(0x%04x) NEXT-IFD=(0x%04x)>", ib.ii.ParentIfdName, ib.ii.IfdName, ib.ifdTagId, len(ib.tags), ib.existingOffset, nextIfdPhrase)
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

    if ib.ii != RootIi {
        log.Panicf("thumbnails can only go into a root Ifd (and only the second one)")
    }

// TODO(dustin): !! Add a test for this function.

    if data == nil || len(data) == 0 {
        log.Panic("thumbnail is empty")
    }

    ib.thumbnailData = data

    ibtvfb := NewIfdBuilderTagValueFromBytes(ib.thumbnailData)
    offsetBt := NewBuilderTag(ib.ii, ThumbnailOffsetTagId, TypeLong, ibtvfb)

    err = ib.Set(offsetBt)
    log.PanicIf(err)

    sizeBt := NewStandardBuilderTag(ib.ii, ThumbnailSizeTagId, ib.byteOrder, []uint32{uint32(len(ib.thumbnailData))})

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
                _, isChildIb := IfdTagNameWithId(currentIb.ii.IfdName, tag.tagId)

                tagName := ""

                // If a normal tag (not a child IFD) get the name.
                if isChildIb == true {
                    tagName = "<Child IFD>"
                } else {
                    it, err := tagIndex.Get(tag.ii, tag.tagId)
                    if log.Is(err, ErrTagNotFound) == true {
                        tagName = "<UNKNOWN>"
                    } else if err != nil {
                        log.Panic(err)
                    } else {
                        tagName = it.Name
                    }
                }

                fmt.Printf("%s  (%d): [%s] %s\n", indent, i, tagName, tag)

                if isChildIb == true {
                    if tag.value.IsIb() == false {
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
                _, isChildIb := IfdTagNameWithId(currentIb.ii.IfdName, tag.tagId)

                if isChildIb == true {
                    if tag.value.IsIb() == false {
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

func (ib *IfdBuilder) dumpToStrings(thisIb *IfdBuilder, prefix string, lines []string) (linesOutput []string) {
    if lines == nil {
        linesOutput = make([]string, 0)
    } else {
        linesOutput = lines
    }

    for i, tag := range thisIb.tags {
        childIfdName := ""
        if tag.value.IsIb() == true {
            childIfdName = tag.value.Ib().ii.IfdName
        }

        line := fmt.Sprintf("<PARENTS=[%s] IFD-NAME=[%s]> IFD-TAG-ID=(0x%04x) CHILD-IFD=[%s] INDEX=(%d) TAG=[0x%04x]", prefix, thisIb.ii.IfdName, thisIb.ifdTagId, childIfdName, i, tag.tagId)
        linesOutput = append(linesOutput, line)

        if tag.value.IsIb() == true {
            childPrefix := ""
            if prefix == "" {
                childPrefix = fmt.Sprintf("%s", thisIb.ii.IfdName)
            } else {
                childPrefix = fmt.Sprintf("%s->%s", prefix, thisIb.ii.IfdName)
            }

            linesOutput = thisIb.dumpToStrings(tag.value.Ib(), childPrefix, linesOutput)
        }
    }

    return linesOutput
}

func (ib *IfdBuilder) DumpToStrings() (lines []string) {
    return ib.dumpToStrings(ib, "", lines)
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
        if log.Is(err, ErrTagEntryNotFound) == true {
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
    } else if log.Is(err, ErrTagEntryNotFound) == true {
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

    it, err := tagIndex.GetWithName(ib.ii, tagName)
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

    if bt.ii == ZeroIi {
        log.Panicf("BuilderTag IfdIdentity is not set: %s", bt)
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

    if bt.value.IsIb() == true {
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
            log.Panicf("child-IFD already added: %v", childIb.ii)
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
        ib.ii,
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

    thumbnailData, err := ifd.Thumbnail()
    if err == nil {
        err = ib.SetThumbnail(thumbnailData)
        log.PanicIf(err)
    } else if log.Is(err, ErrNoThumbnail) == false {
        log.Panic(err)
    }

    for i, ite := range ifd.Entries {
        if ite.TagId == ThumbnailOffsetTagId || ite.TagId == ThumbnailSizeTagId {
            // These will be added on-the-fly when we encode.
            continue
        }

        if excludeTagIds != nil && len(excludeTagIds) > 0 {
            found := false
            for _, excludedTagId := range excludeTagIds {
                if excludedTagId == ite.TagId {
                    found = true
                }
            }

            if found == true {
                continue
            }
        }

        if includeTagIds != nil && len(includeTagIds) > 0 {
            // Whether or not there was a list of excludes, if there is a list
            // of includes than the current tag has to be in it.

            found := false
            for _, includedTagId := range includeTagIds {
                if includedTagId == ite.TagId {
                    found = true
                    break
                }
            }

            if found == false {
                continue
            }
        }

        var bt *BuilderTag

        if ite.ChildIfdName != "" {
            // If we want to add an IFD tag, we'll have to build it first and
            // *then* add it via a different method.

            if itevr == nil {
                // We don't have any ability to resolve the structure of the
                // child-IFD. Just install it as a normal tag rather than a
                // fully-structured child-IFD. We're going to blank the value,
                // though, since its original offset will no longer be valid
                // (nor does it matter since this is just a temporary
                // placeholder, in this situation).
                value := NewIfdBuilderTagValueFromBytes([]byte{0, 0, 0, 0})
                bt = NewBuilderTag(ite.Ii, ite.TagId, ite.TagType, value)
            } else {
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

                    log.Panicf("could not find child IFD for child ITE: II=[%s] TAG-ID=(0x%04x) CURRENT-TAG-POSITION=(%d) CHILDREN=%v", ite.Ii, ite.TagId, i, childTagIds)
                }

                childIb := NewIfdBuilderFromExistingChain(childIfd, itevr)
                bt = ib.NewBuilderTagFromBuilder(childIb)
            }
        } else {
            var value *IfdBuilderTagValue

            if itevr == nil {
                // rawValueOffsetCopy is our own private copy of the original data.
                // It should always be four-bytes, but just copy whatever there is.
                rawValueOffsetCopy := make([]byte, len(ite.RawValueOffset))
                copy(rawValueOffsetCopy, ite.RawValueOffset)

                value = NewIfdBuilderTagValueFromBytes(rawValueOffsetCopy)
            } else {
                var err error


// TODO(dustin): !! Not correct. If we're adding from existing and it's an unknown-type tag that we can't parse, we're just going to be seting the placeholder even though there's nothing stopping us from just taking the raw bytes (other than some design decisions that we'd have to make in order to do this).


                valueBytes, err := itevr.ValueBytes(ite)
                if err != nil {
                    if log.Is(err, ErrUnhandledUnknownTypedTag) == true {
                        valueBytes = []byte(UnparseableUnknownTagValuePlaceholder)
                    } else {
                        log.Panic(err)
                    }
                }

                value = NewIfdBuilderTagValueFromBytes(valueBytes)
            }

            bt = NewBuilderTag(ifd.Ii, ite.TagId, ite.TagType, value)
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

    bt := NewStandardBuilderTag(ib.ii, tagId, ib.byteOrder, value)

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

    bt := NewStandardBuilderTag(ib.ii, tagId, ib.byteOrder, value)

    i, err := ib.Find(tagId)
    log.PanicIf(err)

    ib.tags[i] = bt

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

    bt := NewStandardBuilderTagWithName(ib.ii, tagName, ib.byteOrder, value)

    err = ib.add(bt)
    log.PanicIf(err)

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

    bt := NewStandardBuilderTagWithName(ib.ii, tagName, ib.byteOrder, value)

    i, err := ib.Find(bt.tagId)
    log.PanicIf(err)

    ib.tags[i] = bt

    return nil
}
