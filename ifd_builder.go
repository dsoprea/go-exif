package exif

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
    ib *IfdBuilder
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


type builderTag struct {
    // ii is the IfdIdentity of the IFD that hosts this tag.
    ii IfdIdentity

    tagId uint16
    typeId uint16

    // value is either a value that can be encoded, an IfdBuilder instance (for
    // child IFDs), or an IfdTagEntry instance representing an existing,
    // previously-stored tag.
    value *IfdBuilderTagValue
}

func NewBuilderTag(ii IfdIdentity, tagId uint16, typeId uint16, value *IfdBuilderTagValue) builderTag {
    return builderTag{
        ii: ii,
        tagId: tagId,
        typeId: typeId,
        value: value,
    }
}

func NewStandardBuilderTag(ii IfdIdentity, tagId uint16, value *IfdBuilderTagValue) builderTag {
    ti := NewTagIndex()

    it, err := ti.Get(ii, tagId)
    log.PanicIf(err)

    return builderTag{
        ii: ii,
        tagId: tagId,
        typeId: it.Type,
        value: value,
    }
}

func NewChildIfdBuilderTag(ii IfdIdentity, tagId uint16, value *IfdBuilderTagValue) builderTag {
    return builderTag{
        ii: ii,
        tagId: tagId,
        typeId: TypeLong,
        value: value,
    }
}

func (bt builderTag) Value() (value *IfdBuilderTagValue) {
    return bt.value
}

func (bt builderTag) String() string {
    valuePhrase := ""

    if bt.value.IsBytes() == true {
        valueBytes := bt.value.Bytes()

        if len(valueBytes) <= 8 {
            valuePhrase = fmt.Sprintf("%v", valueBytes)
        } else {
            valuePhrase = fmt.Sprintf("%v...", valueBytes[:8])
        }
    } else {
        valuePhrase = fmt.Sprintf("%v", bt.value.Ib())
    }

    return fmt.Sprintf("BuilderTag<TAG-ID=(0x%02x) IFD=[%s] VALUE=[%v]>", bt.tagId, bt.ii, valuePhrase)
}

// NewStandardBuilderTagFromConfig constructs a `builderTag` instance. The type
// is looked up. `ii` is the type of IFD that owns this tag.
func NewStandardBuilderTagFromConfig(ii IfdIdentity, tagId uint16, byteOrder binary.ByteOrder, value interface{}) builderTag {
    ti := NewTagIndex()

    it, err := ti.Get(ii, tagId)
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

// NewStandardBuilderTagFromConfig constructs a `builderTag` instance. The type is
// explicitly provided. `ii` is the type of IFD that owns this tag.
func NewBuilderTagFromConfig(ii IfdIdentity, tagId uint16, typeId uint16, byteOrder binary.ByteOrder, value interface{}) builderTag {
    tt := NewTagType(typeId, byteOrder)

    ve := NewValueEncoder(byteOrder)

    var ed EncodedData
    if typeId == TypeUndefined {
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

// NewStandardBuilderTagFromConfigWithName allows us to easily generate solid, consistent tags
// for testing with. `ii` is the type of IFD that owns this tag. This can not be
// an IFD (IFDs are not associated with standardized, official names).
func NewStandardBuilderTagFromConfigWithName(ii IfdIdentity, tagName string, byteOrder binary.ByteOrder, value interface{}) builderTag {
    ti := NewTagIndex()

    it, err := ti.GetWithName(ii, tagName)
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
    tags []builderTag

    // existingOffset will be the offset that this IFD is currently found at if
    // it represents an IFD that has previously been stored (or 0 if not).
    existingOffset uint32

    // nextIb represents the next link if we're chaining to another.
    nextIb *IfdBuilder
}

func NewIfdBuilder(ii IfdIdentity, byteOrder binary.ByteOrder) (ib *IfdBuilder) {
    ifdTagId, _ := IfdTagIdWithIdentity(ii)

    ib = &IfdBuilder{
        // ii describes the current IFD and its parent.
        ii: ii,

        // ifdTagId is empty unless it's a child-IFD.
        ifdTagId: ifdTagId,

        byteOrder: byteOrder,
        tags: make([]builderTag, 0),
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
        ii: ii,
        ifdTagId: ifdTagId,
        byteOrder: ifd.ByteOrder,
        existingOffset: ifd.Offset,
    }

    return ib
}

// NewIfdBuilderFromExistingChain creates a chain of IB instances from an
// IFD chain generated from real data.
func NewIfdBuilderFromExistingChain(rootIfd *Ifd, exifData []byte) (firstIb *IfdBuilder) {
    itevr := NewIfdTagEntryValueResolver(exifData, rootIfd.ByteOrder)

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

        // Any child IFDs will still not be copied. Do that now.

        for _, childIfd := range thisExistingIfd.Children {
            childIb := NewIfdBuilderFromExistingChain(childIfd, exifData)

            err = newIb.AddChildIb(childIb)
            log.PanicIf(err)
        }

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

    return fmt.Sprintf("IfdBuilder<PARENT-IFD=[%s] IFD=[%s] TAG-ID=(0x%02x) COUNT=(%d) OFF=(0x%04x) NEXT-IFD=(0x%04x)>", ib.ii.ParentIfdName, ib.ii.IfdName, ib.ifdTagId, len(ib.tags), ib.existingOffset, nextIfdPhrase)
}

func (ib *IfdBuilder) Tags() (tags []builderTag) {
    return ib.tags
}

func (ib *IfdBuilder) printTagTree(levels int) {
    indent := strings.Repeat(" ", levels * 2)

    ti := NewTagIndex()
    i := 0
    for currentIb := ib; currentIb != nil; currentIb = currentIb.nextIb {
        prefix := " "
        if i > 0 {
            prefix = ">"
        }

        if levels == 0 {
            fmt.Printf("%s%sIFD: %s\n", indent, prefix, currentIb)
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
                    it, err := ti.Get(tag.ii, tag.tagId)
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
                        log.Panicf("tag-ID (0x%02x) is an IFD but the tag value is not an IB instance: %v", tag.tagId, tag)
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
    indent := strings.Repeat(" ", levels * 2)

    i := 0
    for currentIb := ib; currentIb != nil; currentIb = currentIb.nextIb {
        prefix := " "
        if i > 0 {
            prefix = ">"
        }

        fmt.Printf("%s%s%s\n", indent, prefix,currentIb)

        if len(currentIb.tags) > 0 {
            for _, tag := range currentIb.tags {
                _, isChildIb := IfdTagNameWithId(currentIb.ii.IfdName, tag.tagId)

                if isChildIb == true {
                    if tag.value.IsIb() == false {
                        log.Panicf("tag-ID (0x%02x) is an IFD but the tag value is not an IB instance: %v", tag.tagId, tag)
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

        line := fmt.Sprintf("<PARENTS=[%s] IFD-NAME=[%s]> IFD-TAG-ID=(0x%02x) CHILD-IFD=[%s] INDEX=(%d) TAG=[0x%02x]", prefix, thisIb.ii.IfdName, thisIb.ifdTagId, childIfdName, i, tag.tagId)
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

    for ; n > 0; {
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

        ib.tags = append(ib.tags[:j], ib.tags[j + 1:]...)
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

func (ib *IfdBuilder) ReplaceAt(position int, bt builderTag) (err error) {
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

func (ib *IfdBuilder) Replace(tagId uint16, bt builderTag) (err error) {
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

func (ib *IfdBuilder) Add(bt builderTag) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if bt.ii == ZeroIi {
        log.Panicf("builderTag IfdIdentity is not set: %s", bt)
    } else if bt.typeId == 0x0 {
        log.Panicf("builderTag type-ID is not set: %s", bt)
    } else if bt.value == nil {
        log.Panicf("builderTag value is not set: %s", bt)
    }

    if bt.value.IsIb() == true {
        log.Panicf("child IfdBuilders must be added via AddChildIb() not Add()")
    }

    ib.tags = append(ib.tags, bt)
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

    value := NewIfdBuilderTagValueFromIfdBuilder(childIb)

    bt := NewChildIfdBuilderTag(
            ib.ii,
            childIb.ifdTagId,
            value)

    ib.tags = append(ib.tags, bt)

    return nil
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

    for _, ite := range ifd.Entries {
        // If we want to add an IFD tag, we'll have to build it first and *then*
        // add it via a different method.
        if ite.ChildIfdName != "" {
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

        var value *IfdBuilderTagValue

        if itevr == nil {
            // rawValueOffsetCopy is our own private copy of the original data.
            // It should always be four-bytes, but just copy whatever there is.
            rawValueOffsetCopy := make([]byte, len(ite.RawValueOffset))
            copy(rawValueOffsetCopy, ite.RawValueOffset)

            value = NewIfdBuilderTagValueFromBytes(rawValueOffsetCopy)
        } else {
            var err error

            valueBytes, err := itevr.ValueBytes(ite)
            if err != nil {
                if log.Is(err, ErrUnhandledUnknownTypedTag) == true {
                    ifdBuilderLogger.Warningf(nil, "Unknown-type tag can't be parsed so it can't be copied to the new IFD.")
                    continue
                }

                log.Panic(err)
            }

            value = NewIfdBuilderTagValueFromBytes(valueBytes)
        }

        bt := NewBuilderTag(ifd.Ii, ite.TagId, ite.TagType, value)

        err := ib.Add(bt)
        log.PanicIf(err)
    }

    return nil
}

// AddFromConfig quickly and easily composes and adds the tag using the
// information already known about a tag. Only works with standard tags.
func (ib *IfdBuilder) AddFromConfig(tagId uint16, value interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    bt := NewStandardBuilderTagFromConfig(ib.ii, tagId, ib.byteOrder, value)

    err = ib.Add(bt)
    log.PanicIf(err)

    return nil
}

// SetFromConfig quickly and easily composes and adds or replaces the tag using
// the information already known about a tag. Only works with standard tags.
func (ib *IfdBuilder) SetFromConfig(tagId uint16, value interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

// TODO(dustin): !! Add test.

    bt := NewStandardBuilderTagFromConfig(ib.ii, tagId, ib.byteOrder, value)

    i, err := ib.Find(tagId)
    log.PanicIf(err)

    ib.tags[i] = bt

    return nil
}

// AddFromConfigWithName quickly and easily composes and adds the tag using the
// information already known about a tag (using the name). Only works with
// standard tags.
func (ib *IfdBuilder) AddFromConfigWithName(tagName string, value interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    bt := NewStandardBuilderTagFromConfigWithName(ib.ii, tagName, ib.byteOrder, value)

    err = ib.Add(bt)
    log.PanicIf(err)

    return nil
}

// SetFromConfigWithName quickly and easily composes and adds or replaces the
// tag using the information already known about a tag (using the name). Only
// works with standard tags.
func (ib *IfdBuilder) SetFromConfigWithName(tagName string, value interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

// TODO(dustin): !! Add test.

    bt := NewStandardBuilderTagFromConfigWithName(ib.ii, tagName, ib.byteOrder, value)

    i, err := ib.Find(bt.tagId)
    log.PanicIf(err)

    ib.tags[i] = bt

    return nil
}
