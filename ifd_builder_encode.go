package exif

import (
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)


type ByteWriter struct {
    b *bytes.Buffer
    byteOrder binary.ByteOrder
}

func NewByteWriter(b *bytes.Buffer, byteOrder binary.ByteOrder) (bw *ByteWriter) {
    return &ByteWriter{
        b: b,
        byteOrder: byteOrder,
    }
}

func (bw ByteWriter) writeAsBytes(value interface{}) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = binary.Write(bw.b, bw.byteOrder, value)
    log.PanicIf(err)

    return nil
}

func (bw ByteWriter) WriteUint32(value uint32) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = bw.writeAsBytes(value)
    log.PanicIf(err)

    return nil
}

func (bw ByteWriter) WriteUint16(value uint16) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = bw.writeAsBytes(value)
    log.PanicIf(err)

    return nil
}

func (bw ByteWriter) WriteFourBytes(value []byte) (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    len_ := len(value)
    if len_ != 4 {
        log.Panicf("value is not four-bytes: (%d)", len_)
    }

    _, err = bw.b.Write(value)
    log.PanicIf(err)

    return nil
}


// ifdOffsetIterator keeps track of where the next IFD should be written by
// keeping track of where the offsets start, the data that has been added, and
// bumping the offset *when* the data is added.
type ifdDataAllocator struct {
    offset uint32
    b bytes.Buffer
}

func newIfdDataAllocator(ifdDataAddressableOffset uint32) *ifdDataAllocator{
    return &ifdDataAllocator{
        offset: ifdDataAddressableOffset,
    }
}

func (ida *ifdDataAllocator) Allocate(value []byte) (offset uint32, err error) {
    _, err = ida.b.Write(value)
    log.PanicIf(err)

    offset = ida.offset
    ida.offset += uint32(len(value))

    return offset, nil
}

func (ida *ifdDataAllocator) NextOffset() uint32 {
    return ida.offset
}

func (ida *ifdDataAllocator) Bytes() []byte {
    return ida.b.Bytes()
}


// IfdByteEncoder converts an IB to raw bytes (for writing) while also figuring
// out all of the allocations and indirection that is required for extended
// data.
type IfdByteEncoder struct {
}

func NewIfdByteEncoder() (ibe *IfdByteEncoder) {
    return new(IfdByteEncoder)
}

func (ibe *IfdByteEncoder) EntrySize() uint32 {
    // Tag-ID + Tag-Type + Unit-Count + Value/Offset.
    return uint32(2 + 2 + 4 + 4)
}

func (ibe *IfdByteEncoder) TableSize(entryCount int) uint32 {
    // Tag-Count + (Entry-Size * Entry-Count) + Next-IFD-Offset.
    return uint32(2) + (ibe.EntrySize() * uint32(entryCount)) + uint32(4)
}

// encodeTagToBytes encodes the given tag to a byte stream. If
// `nextIfdOffsetToWrite` is more than (0), recurse into child IFDs
// (`nextIfdOffsetToWrite` is required in order for them to know where the its
// IFD data will be written, in order for them to know the offset of where
// their allocated-data block will start, which follows right behind).
func (ibe *IfdByteEncoder) encodeTagToBytes(ib *IfdBuilder, bt *builderTag, bw *ByteWriter, ida *ifdDataAllocator, nextIfdOffsetToWrite uint32) (childIfdBlock []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    // Write tag-ID.
    err = bw.WriteUint16(bt.tagId)
    log.PanicIf(err)

    // Works for both values and child IFDs (which have an official size of
    // LONG).
    err = bw.WriteUint16(bt.typeId)
    log.PanicIf(err)

    // Write unit-count.

    if bt.value.IsBytes() == true {
        effectiveType := bt.typeId
        if bt.typeId == TypeUndefined {
            effectiveType = TypeByte
        }

        // It's a non-unknown value.Calculate the count of values of
        // the type that we're writing and the raw bytes for the whole list.

        typeSize := uint32(TagTypeSize(effectiveType))

        valueBytes := bt.value.Bytes()

        len_ := len(valueBytes)
        unitCount := uint32(len_) / typeSize
        remainder := uint32(len_) % typeSize

        if remainder > 0 {
            log.Panicf("tag value of (%d) bytes not evenly divisible by type-size (%d)", len_, typeSize)
        }

        err = bw.WriteUint32(unitCount)
        log.PanicIf(err)

        // Write four-byte value/offset.

        if len_ > 4 {
            offset, err := ida.Allocate(valueBytes)
            log.PanicIf(err)

            err = bw.WriteUint32(offset)
            log.PanicIf(err)
        } else {
            fourBytes := make([]byte, 4)
            copy(fourBytes, valueBytes)

            err = bw.WriteFourBytes(fourBytes)
            log.PanicIf(err)
        }
    } else {
        if bt.value.IsIb() == false {
            log.Panicf("tag value is not a byte-slice but also not a child IB: %v", bt)
        }

        // Write unit-count (one LONG representing one offset).
        err = bw.WriteUint32(1)
        log.PanicIf(err)

        if nextIfdOffsetToWrite > 0 {
            var err error

            // Create the block of IFD data and everything it requires.
            childIfdBlock, err = ibe.encodeAndAttachIfd(bt.value.Ib(), nextIfdOffsetToWrite)
            log.PanicIf(err)

            // Use the next-IFD offset for it. The IFD will actually get
            // attached after we return.
            err = bw.WriteUint32(nextIfdOffsetToWrite)
            log.PanicIf(err)

        } else {
            // No child-IFDs are to be allocated. Finish the entry with a NULL
            // pointer.

            err = bw.WriteUint32(0)
            log.PanicIf(err)
        }
    }

    return childIfdBlock, nil
}

// encodeIfdToBytes encodes the given IB to a byte-slice. We are given the
// offset at which this IFD will be written. This method is used called both to
// pre-determine how big the table is going to be (so that we can calculate the
// address to allocate data at) as well as to write the final table.
//
// It is necessary to fully realize the table in order to predetermine its size
// because it is not enough to know the size of the table: If there are child
// IFDs, we will not be able to allocate them without first knowing how much
// data we need to allocate for the current IFD.
func (ibe *IfdByteEncoder) encodeIfdToBytes(ib *IfdBuilder, ifdAddressableOffset uint32, nextIfdOffsetToWrite uint32, setNextIb bool) (data []byte, tableSize uint32, dataSize uint32, childIfdSizes []uint32, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    tableSize = ibe.TableSize(len(ib.tags))

    b := new(bytes.Buffer)
    bw := NewByteWriter(b, ib.byteOrder)

    // Write tag count.
    err = bw.WriteUint16(uint16(len(ib.tags)))
    log.PanicIf(err)

    ida := newIfdDataAllocator(ifdAddressableOffset)

    childIfdBlocks := make([][]byte, 0)

    // Write raw bytes for each tag entry. Allocate larger data to be referred
    // to in the follow-up data-block as required. Any "unknown"-byte tags that
    // we can't parse will not be present here (using AddTagsFromExisting(), at
    // least).
    for _, bt := range ib.tags {
        childIfdBlock, err := ibe.encodeTagToBytes(ib, &bt, bw, ida, nextIfdOffsetToWrite)
        log.PanicIf(err)

        if childIfdBlock != nil {
            if nextIfdOffsetToWrite == 0 {
                log.Panicf("no IFD offset provided for child-IFDs; no new child-IFDs permitted")
            }

            nextIfdOffsetToWrite += uint32(len(childIfdBlock))
            childIfdBlocks = append(childIfdBlocks, childIfdBlock)
        }
    }

    dataBytes := ida.Bytes()
    dataSize = uint32(len(dataBytes))

    childIfdSizes = make([]uint32, len(childIfdBlocks))
    childIfdsTotalSize := uint32(0)
    for i, childIfdBlock := range childIfdBlocks {
        len_ := uint32(len(childIfdBlock))
        childIfdSizes[i] = len_
        childIfdsTotalSize += len_
    }

    // Set the link from this IFD to the next IFD that will be written in the
    // next cycle.
    if setNextIb == true {
        nextIfdOffsetToWrite += tableSize + dataSize + childIfdsTotalSize
    } else {
        nextIfdOffsetToWrite = 0
    }

    // Write address of next IFD in chain.
    err = bw.WriteUint32(nextIfdOffsetToWrite)
    log.PanicIf(err)

    _, err = b.Write(dataBytes)
    log.PanicIf(err)

    // Append any child IFD blocks after our table and data blocks. These IFDs
    // were equipped with the appropriate offset information so it's expected
    // that all offsets referred to by these will be correct.
    //
    // Note that child-IFDs are append after the current IFD and before the
    // next IFD, as opposed to the root IFDs, which are chained together but
    // will be interrupted by these child-IFDs (which is expected, per the
    // standard).

    for _, childIfdBlock := range childIfdBlocks {
        _, err = b.Write(childIfdBlock)
        log.PanicIf(err)
    }

    return b.Bytes(), tableSize, dataSize, childIfdSizes, nil
}

// encodeAndAttachIfd is a reentrant function that processes the IFD chain.
func (ibe *IfdByteEncoder) encodeAndAttachIfd(ib *IfdBuilder, ifdAddressableOffset uint32) (data []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if len(ib.tags) == 0 {
        log.Panicf("trying to encode an IfdBuilder that doesn't have any tags")
    }

    b := new(bytes.Buffer)

    nextIfdOffsetToWrite := uint32(0)
    for thisIb := ib; thisIb != nil; thisIb = thisIb.nextIb {
        // Do a dry-run in order to pre-determine its size requirement.

        _, tableSize, allocatedDataSize, _, err := ibe.encodeIfdToBytes(ib, ifdAddressableOffset, 0, false)
        log.PanicIf(err)

        addressableOffset := ifdAddressableOffset + tableSize
        nextIfdOffsetToWrite = addressableOffset + allocatedDataSize

        // Write our IFD as well as any child-IFDs (now that we know the offset
        // where new IFDs and their data will be allocated).

        setNextIb := thisIb.nextIb != nil

        tableAndAllocated, tableSize, allocatedDataSize, _, err := ibe.encodeIfdToBytes(ib, addressableOffset, nextIfdOffsetToWrite, setNextIb)
        log.PanicIf(err)

        if len(tableAndAllocated) != int(tableSize + allocatedDataSize) {
            log.Panicf("IFD table and data is not a consistent size: (%d) != (%d)", len(tableAndAllocated), tableSize + allocatedDataSize)
        }

        _, err = b.Write(tableAndAllocated)
        log.PanicIf(err)

        // This will include the child-IFDs, as well. This will actually advance the offset for our next loop.
        ifdAddressableOffset = ifdAddressableOffset + uint32(tableSize + allocatedDataSize)
    }

    return b.Bytes(), nil
}

// EncodeToExifPayload is the base encoding step that transcribes the entire IB
// structure to its on-disk layout.
func (ibe *IfdByteEncoder) EncodeToExifPayload(ib *IfdBuilder) (data []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    data, err = ibe.encodeAndAttachIfd(ib, ExifDefaultFirstIfdOffset)
    log.PanicIf(err)

    return data, nil
}

// EncodeToExif calls EncodeToExifPayload and then packages the result into a
// complete EXIF block.
func (ibe *IfdByteEncoder) EncodeToExif(ib *IfdBuilder) (data []byte, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    encodedIfds, err := ibe.EncodeToExifPayload(ib)
    log.PanicIf(err)

    // Wrap the IFD in a formal EXIF block.

    b := new(bytes.Buffer)

    headerBytes, err := BuildExifHeader(EncodeDefaultByteOrder, ExifDefaultFirstIfdOffset)
    log.PanicIf(err)

    _, err = b.Write(headerBytes)
    log.PanicIf(err)

    _, err = b.Write(encodedIfds)
    log.PanicIf(err)

    return b.Bytes(), nil
}
