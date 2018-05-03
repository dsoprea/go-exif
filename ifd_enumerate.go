package exif

import (
    "bytes"
    "fmt"
    "strings"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    ifdEnumerateLogger = log.NewLogger("exifjpeg.ifd")
)


// IfdTagEnumerator knows how to decode an IFD and all of the tags it
// describes.
//
// The IFDs and the actual values can float throughout the EXIF block, but the
// IFD itself is just a minor header followed by a set of repeating,
// statically-sized records. So, the tags (though notnecessarily their values)
// are fairly simple to enumerate.
type IfdTagEnumerator struct {
    byteOrder binary.ByteOrder
    addressableData []byte
    ifdOffset uint32
    buffer *bytes.Buffer
}

func NewIfdTagEnumerator(addressableData []byte, byteOrder binary.ByteOrder, ifdOffset uint32) (ite *IfdTagEnumerator) {
    ite = &IfdTagEnumerator{
        addressableData: addressableData,
        byteOrder: byteOrder,
        buffer: bytes.NewBuffer(addressableData[ifdOffset:]),
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
    exifData []byte
    buffer *bytes.Buffer
    byteOrder binary.ByteOrder
    currentOffset uint32
}

func NewIfdEnumerate(exifData []byte, byteOrder binary.ByteOrder) *IfdEnumerate {
    // Make it obvious what data we expect and when we don't get it.
    if IsExif(exifData) == false {
        log.Panicf("not exif data")
    }

    return &IfdEnumerate{
        exifData: exifData,
        buffer: bytes.NewBuffer(exifData),
        byteOrder: byteOrder,
    }
}

// ValueContext describes all of the parameters required to find and extract
// the actual tag value.
type ValueContext struct {
    UnitCount uint32
    ValueOffset uint32
    RawValueOffset []byte
    AddressableData []byte
}

func (ie *IfdEnumerate) getTagEnumerator(ifdOffset uint32) (ite *IfdTagEnumerator) {
    ite = NewIfdTagEnumerator(
            ie.exifData[ExifAddressableAreaStart:],
            ie.byteOrder,
            ifdOffset)

    return ite
}

func (ie *IfdEnumerate) parseTag(ii IfdIdentity, tagIndex int, ite *IfdTagEnumerator) (tag *IfdTagEntry, err error) {
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
        Ii: ii,
        TagId: tagId,
        TagIndex: tagIndex,
        TagType: tagType,
        UnitCount: unitCount,
        ValueOffset: valueOffset,
        RawValueOffset: rawValueOffset,
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

// TagVisitor is an optional callback that can get hit for every tag we parse
// through. `addressableData` is the byte array startign after the EXIF header
// (where the offsets of all IFDs and values are calculated from).
type TagVisitor func(ii IfdIdentity, ifdIndex int, tagId uint16, tagType TagType, valueContext ValueContext) (err error)

// ParseIfd decodes the IFD block that we're currently sitting on the first
// byte of.
func (ie *IfdEnumerate) ParseIfd(ii IfdIdentity, ifdIndex int, ite *IfdTagEnumerator, visitor TagVisitor, doDescend bool) (nextIfdOffset uint32, entries []*IfdTagEntry, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    tagCount, _, err := ite.getUint16()
    log.PanicIf(err)

    ifdEnumerateLogger.Debugf(nil, "Current IFD tag-count: (%d)", tagCount)

    entries = make([]*IfdTagEntry, tagCount)

    for i := 0; i < int(tagCount); i++ {
        tag, err := ie.parseTag(ii, i, ite)
        log.PanicIf(err)

        if visitor != nil {
            tt := NewTagType(tag.TagType, ie.byteOrder)

            vc := ValueContext{
                UnitCount: tag.UnitCount,
                ValueOffset: tag.ValueOffset,
                RawValueOffset: tag.RawValueOffset,
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

            err := ie.scan(childIi, tag.ValueOffset, visitor)
            log.PanicIf(err)
        }

        entries[i] = tag
    }

    nextIfdOffset, _, err = ite.getUint32()
    log.PanicIf(err)

    ifdEnumerateLogger.Debugf(nil, "Next IFD at offset: (%08x)", nextIfdOffset)

    return nextIfdOffset, entries, nil
}

// Scan enumerates the different EXIF blocks (called IFDs).
func (ie *IfdEnumerate) scan(ii IfdIdentity, ifdOffset uint32, visitor TagVisitor) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    for ifdIndex := 0;; ifdIndex++ {
        ifdEnumerateLogger.Debugf(nil, "Parsing IFD [%s] (%d) at offset (%04x).", ii.IfdName, ifdIndex, ifdOffset)
        ite := ie.getTagEnumerator(ifdOffset)

        nextIfdOffset, _, err := ie.ParseIfd(ii, ifdIndex, ite, visitor, true)
        log.PanicIf(err)

        if nextIfdOffset == 0 {
            break
        }

        ifdOffset = nextIfdOffset
    }

    return nil
}

// Scan enumerates the different EXIF blocks (called IFDs).
func (ie *IfdEnumerate) Scan(ifdOffset uint32, visitor TagVisitor) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ii, _ := IfdIdOrFail("", IfdStandard)

    err = ie.scan(ii, ifdOffset, visitor)
    log.PanicIf(err)

    return nil
}


type Ifd struct {
    ByteOrder binary.ByteOrder

    Id int
    ParentIfd *Ifd
    Name string
    Index int
    Offset uint32

    Entries []*IfdTagEntry
    EntriesByTagId map[uint16][]*IfdTagEntry

    Children []*Ifd
    NextIfdOffset uint32
    NextIfd *Ifd
}

// FindTagWithId returns a list of tags (usually just zero or one) that match
// the given tag ID. This is efficient.
func (ifd Ifd) FindTagWithId(tagId uint16) (results []*IfdTagEntry, err error) {
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
func (ifd Ifd) FindTagWithName(tagName string) (results []*IfdTagEntry, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ti := NewTagIndex()

    ii := ifd.Identity()
    it, err := ti.GetWithName(ii, tagName)
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

    return fmt.Sprintf("IFD<ID=(%d) N=[%s] IDX=(%d) OFF=(0x%04x) COUNT=(%d) CHILDREN=(%d) PARENT=(0x%04x) NEXT-IFD=(0x%04x)", ifd.Id, ifd.Name, ifd.Index, ifd.Offset, len(ifd.Entries), len(ifd.Children), parentOffset, ifd.NextIfdOffset)
}

func (ifd Ifd) Identity() IfdIdentity {
    parentIfdName := ""
    if ifd.ParentIfd != nil {
        parentIfdName = ifd.ParentIfd.Name
    }

// TODO(dustin): !! We should be checking using the parent ID, not the parent name.
    ii, _ := IfdIdOrFail(parentIfdName, ifd.Name)

    return ii
}

func (ifd Ifd) printNode(level int, nextLink bool) {
    indent := strings.Repeat(" ", level * 2)

    prefix := " "
    if nextLink {
        prefix = ">"
    }

    fmt.Printf("%s%s%s\n", indent, prefix, ifd)

    for _, childIfd := range ifd.Children {
        childIfd.printNode(level + 1, false)
    }

    if ifd.NextIfd != nil {
        ifd.NextIfd.printNode(level, true)
    }
}

func (ifd Ifd) PrintTree() {
    ifd.printNode(0, false)
}


type QueuedIfd struct {
    Ii IfdIdentity
    Index int
    Offset uint32
    Parent *Ifd
}


type IfdIndex struct {
    RootIfd *Ifd
    Ifds []*Ifd
    Tree map[int]*Ifd
    Lookup map[IfdIdentity][]*Ifd
}


// Scan enumerates the different EXIF blocks (called IFDs).
func (ie *IfdEnumerate) Collect(rootIfdOffset uint32) (index IfdIndex, err error) {
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
            Ii: RootIi,
            Index: 0,
            Offset: rootIfdOffset,
        },
    }

    edges := make(map[uint32]*Ifd)

    for {
        if len(queue) == 0 {
            break
        }

        ii := queue[0].Ii
        name := ii.IfdName
        index := queue[0].Index
        offset := queue[0].Offset
        parentIfd := queue[0].Parent

        queue = queue[1:]

        ifdEnumerateLogger.Debugf(nil, "Parsing IFD [%s] (%d) at offset (%04x).", ii.IfdName, index, offset)
        ite := ie.getTagEnumerator(offset)

        nextIfdOffset, entries, err := ie.ParseIfd(ii, index, ite, nil, false)
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

        ifd := Ifd{
            ByteOrder: ie.byteOrder,
            Id: id,
            ParentIfd: parentIfd,
            Name: name,
            Index: index,
            Offset: offset,
            Entries: entries,
            EntriesByTagId: entriesByTagId,
            Children: make([]*Ifd, 0),
            NextIfdOffset: nextIfdOffset,
        }

        // Add ourselves to a big list of IFDs.
        ifds = append(ifds, &ifd)

        // Install ourselves into a by-id lookup table (keys are unique).
        tree[id] = &ifd

        // Install into by-name buckets.

        if list_, found := lookup[ii]; found == true {
            lookup[ii] = append(list_, &ifd)
        } else {
            list_ = make([]*Ifd, 1)
            list_[0] = &ifd

            lookup[ii] = list_
        }

        // Add a link from the previous IFD in the chain to us.
        if previousIfd, found := edges[offset]; found == true {
            previousIfd.NextIfd = &ifd
        }

        // Attach as a child to our parent (where we appeared as a tag in
        // that IFD).
        if parentIfd != nil {
            parentIfd.Children = append(parentIfd.Children, &ifd)
        }

        // Determine if any of our entries is a child IFD and queue it.
        for _, entry := range entries {
            if entry.ChildIfdName == "" {
                continue
            }

            childId := IfdIdentity{
                ParentIfdName: name,
                IfdName: entry.ChildIfdName,
            }

            qi := QueuedIfd{
                Ii: childId,
                Index: 0,
                Offset: entry.ValueOffset,
                Parent: &ifd,
            }

            queue = append(queue, qi)
        }

        // If there's another IFD in the chain.
        if nextIfdOffset != 0 {
            // Allow the next link to know what the previous link was.
            edges[nextIfdOffset] = &ifd

            qi := QueuedIfd{
                Ii: ii,
                Index: index + 1,
                Offset: nextIfdOffset,
            }

            queue = append(queue, qi)
        }
    }

    index.RootIfd = tree[0]
    index.Ifds = ifds
    index.Tree = tree
    index.Lookup = lookup

    return index, nil
}

// ParseOneIfd is a hack to use an IE to parse a raw IFD block. Can be used for
// testing.
func ParseOneIfd(ii IfdIdentity, byteOrder binary.ByteOrder, ifdBlock []byte, visitor TagVisitor) (nextIfdOffset uint32, entries []*IfdTagEntry, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ie := &IfdEnumerate{
        byteOrder: byteOrder,
    }

    ite := NewIfdTagEnumerator(ifdBlock, byteOrder, 0)

    nextIfdOffset, entries, err = ie.ParseIfd(ii, 0, ite, visitor, true)
    log.PanicIf(err)

    return nextIfdOffset, entries, nil
}

// ParseOneTag is a hack to use an IE to parse a raw tag block.
func ParseOneTag(ii IfdIdentity, byteOrder binary.ByteOrder, tagBlock []byte) (tag *IfdTagEntry, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ie := &IfdEnumerate{
        byteOrder: byteOrder,
    }

    ite := NewIfdTagEnumerator(tagBlock, byteOrder, 0)

    tag, err = ie.parseTag(ii, 0, ite)
    log.PanicIf(err)

    return tag, nil
}
