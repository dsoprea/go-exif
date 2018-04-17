package exif

import (
    "errors"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

var (
    ifdBuilderLogger = log.NewLogger("exif.ifd_builder")
)

var (
    ErrTagEntryNotFound = errors.New("tag entry not found")
)


// TODO(dustin): !! Make sure we either replace existing IFDs or validate that the IFD doesn't already exist.


type builderTag struct {
    ifdName string
    tagId uint16

    // value is either a value that can be encoded, an IfdBuilder instance (for
    // child IFDs), or an IfdTagEntry instance representing an existing,
    // previously-stored tag.
    value interface{}
}

type IfdBuilder struct {
    ifdName string

    // ifdTagId will be non-zero if we're a child IFD.
    ifdTagId uint16

    byteOrder binary.ByteOrder

    // Includes both normal tags and IFD tags (which point to child IFDs).
    tags []builderTag

    // existingOffset will be the offset that this IFD is currently found at if
    // it represents an IFD that has previously been stored (or 0 if not).
    existingOffset uint32

    // nextIfd represents the next link if we're chaining to another.
    nextIfd *IfdBuilder
}

func NewIfdBuilder(ifdName string, byteOrder binary.ByteOrder) (ib *IfdBuilder) {
    ib = &IfdBuilder{
        ifdName: ifdName,
        ifdTagId: IfdTagIds[ifdName],
        byteOrder: byteOrder,
        tags: make([]builderTag, 0),
    }

    return ib
}

func NewIfdBuilderWithExistingIfd(ifd *Ifd, byteOrder binary.ByteOrder) (ib *IfdBuilder) {
    ib = &IfdBuilder{
        ifdName: ifd.Name,
        byteOrder: byteOrder,
        existingOffset: ifd.Offset,
    }

    return ib
}

func (ib *IfdBuilder) SetNextIfd(nextIfd *IfdBuilder) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ib.nextIfd = nextIfd

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

func (ib *IfdBuilder) DeleteAll(tagId uint16) (err error) {
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

// TODO(dustin): !! Switch to producing bytes immediately so that they're validated.

func (ib *IfdBuilder) Add(bt builderTag) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    ib.tags = append(ib.tags, bt)
    return nil
}

func (ib *IfdBuilder) AddChildIfd(childIfd *IfdBuilder) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

// TODO(dustin): !! We might not want to take an actual IfdBuilder instance, as
//                  these are mutable in nature (unless we definitely want to
//                  allow them to tbe chnaged right up until they're actually
//                  written). We might be better with a final, immutable tag
//                  container insted.

    if childIfd.ifdTagId == 0 {
        log.Panicf("IFD [%s] can not be used as a child IFD (not associated with a tag-ID)")
    }

    bt := builderTag{
        ifdName: childIfd.ifdName,
        tagId: childIfd.ifdTagId,
        value: childIfd,
    }

    ib.Add(bt)

    return nil
}

// AddTagsFromExisting does a verbatim copy of the entries in `ifd` to this
// builder. It excludes child IFDs. This must be added explicitly via
// `AddChildIfd()`.
func (ib *IfdBuilder) AddTagsFromExisting(ifd *Ifd, includeTagIds []uint16, excludeTagIds []uint16) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()


// Notes: This is used to update existing IFDs (by constructing a new IFD with existing information).
// - How to handle the existing allocation? Obviously this will be an update
//   operation, we should try and re-use the current space.
// - Inevitably, there will be fragmentation as IFDs are changed. We might not
//   be able to avoid reallocation.
//   - !! We'll potentially have to update every recorded tag and IFD offset.
//     - We might just have to refuse to allow updates if we encountered any
//       unmanageable tags (we'll definitely have to finish adding support for
//       the well-known ones).
//
// - An IfdEnumerator might not be the right type of argument, here. It actively
//   reads from a file and is not just a static container.
//   - We might want to create a static-container type that can populate from
//     an IfdEnumerator and then be read and re-read (like an IEnumerable vs IList).



    for _, tag := range ifd.Entries {
        // If we want to add an IFD tag, we'll have to build it first and *then*
        // add it via a different method.
        if tag.IfdName != "" {
            continue
        }

        if excludeTagIds != nil && len(excludeTagIds) > 0 {
            for _, excludedTagId := range excludeTagIds {
                if excludedTagId == tag.TagId {
                    continue
                }
            }
        }

        if includeTagIds != nil && len(includeTagIds) > 0 {
            // Whether or not there was a list of excludes, if there is a list
            // of includes than the current tag has to be in it.

            found := false
            for _, includedTagId := range includeTagIds {
                if includedTagId == tag.TagId {
                    found = true
                    break
                }
            }

            if found == false {
                continue
            }
        }

        bt := builderTag{
            tagId: tag.TagId,

// TODO(dustin): !! For right now, a IfdTagEntry instance will mean that the value will have to be inherited/copied from an existing offset.
            value: tag,
        }

        err := ib.Add(bt)
        log.PanicIf(err)
    }

    return nil
}
