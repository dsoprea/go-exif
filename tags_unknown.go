package exif

import (
    "fmt"
    "strings"
    "bytes"

    "encoding/binary"

    "github.com/dsoprea/go-logging"
)

const (
    TagUnknownType_9298_UserComment_Encoding_ASCII = iota
    TagUnknownType_9298_UserComment_Encoding_JIS = iota
    TagUnknownType_9298_UserComment_Encoding_UNICODE = iota
    TagUnknownType_9298_UserComment_Encoding_UNDEFINED = iota
)

const (
    TagUnknownType_9101_ComponentsConfiguration_Channel_Y = 0x1
    TagUnknownType_9101_ComponentsConfiguration_Channel_Cb = 0x2
    TagUnknownType_9101_ComponentsConfiguration_Channel_Cr = 0x3
    TagUnknownType_9101_ComponentsConfiguration_Channel_R = 0x4
    TagUnknownType_9101_ComponentsConfiguration_Channel_G = 0x5
    TagUnknownType_9101_ComponentsConfiguration_Channel_B = 0x6
)

const (
    TagUnknownType_9101_ComponentsConfiguration_OTHER = iota
    TagUnknownType_9101_ComponentsConfiguration_RGB = iota
    TagUnknownType_9101_ComponentsConfiguration_YCBCR = iota
)

var (
    TagUnknownType_9298_UserComment_Encoding_Names = map[int]string {
        TagUnknownType_9298_UserComment_Encoding_ASCII: "ASCII",
        TagUnknownType_9298_UserComment_Encoding_JIS: "JIS",
        TagUnknownType_9298_UserComment_Encoding_UNICODE: "UNICODE",
        TagUnknownType_9298_UserComment_Encoding_UNDEFINED: "UNDEFINED",
    }

    TagUnknownType_9298_UserComment_Encodings = map[int][]byte {
        TagUnknownType_9298_UserComment_Encoding_ASCII:
            []byte { 'A', 'S', 'C', 'I', 'I', 0, 0, 0 },
        TagUnknownType_9298_UserComment_Encoding_JIS:
            []byte { 'J', 'I', 'S', 0, 0, 0, 0, 0 },
        TagUnknownType_9298_UserComment_Encoding_UNICODE:
            []byte { 'U', 'n', 'i', 'c', 'o', 'd', 'e', 0 },
        TagUnknownType_9298_UserComment_Encoding_UNDEFINED:
            []byte { 0, 0, 0, 0, 0, 0, 0, 0 },
    }

    TagUnknownType_9101_ComponentsConfiguration_Names = map[int]string {
        TagUnknownType_9101_ComponentsConfiguration_OTHER: "OTHER",
        TagUnknownType_9101_ComponentsConfiguration_RGB: "RGB",
        TagUnknownType_9101_ComponentsConfiguration_YCBCR: "YCBCR",
    }

    TagUnknownType_9101_ComponentsConfiguration_Configurations = map[int][]byte {
        TagUnknownType_9101_ComponentsConfiguration_RGB: []byte {
            TagUnknownType_9101_ComponentsConfiguration_Channel_R,
            TagUnknownType_9101_ComponentsConfiguration_Channel_G,
            TagUnknownType_9101_ComponentsConfiguration_Channel_B,
            0,
        },

        TagUnknownType_9101_ComponentsConfiguration_YCBCR: []byte {
            TagUnknownType_9101_ComponentsConfiguration_Channel_Y,
            TagUnknownType_9101_ComponentsConfiguration_Channel_Cb,
            TagUnknownType_9101_ComponentsConfiguration_Channel_Cr,
            0,
        },
    }
)


type UnknownTagValue interface {
    ValueBytes() ([]byte, error)
}


type TagUnknownType_GeneralString string

func (gs TagUnknownType_GeneralString) ValueBytes() (value []byte, err error) {
    return []byte(gs), nil
}


type TagUnknownType_9298_UserComment struct {
    EncodingType int
    EncodingBytes []byte
}

func (uc TagUnknownType_9298_UserComment) String() string {
    return fmt.Sprintf("UserComment<ENCODING=[%s] V=%v>", TagUnknownType_9298_UserComment_Encoding_Names[uc.EncodingType], uc.EncodingBytes)
}

func (uc TagUnknownType_9298_UserComment) ValueBytes() (value []byte, err error) {
    encodingTypeBytes, found := TagUnknownType_9298_UserComment_Encodings[uc.EncodingType]
    if found == false {
        log.Panicf("encoding-type not valid for unknown-type tag 9298 (UserComment): (%d)", uc.EncodingType)
    }

    value = make([]byte, len(uc.EncodingBytes) + 8)
    copy(value[:8], encodingTypeBytes)

// TODO(dustin): !! With undefined-encoded comments, we always make this empty. However, it comes in with a set of zero bytes. Is there a problem if we send it out with just the encoding bytes?
    copy(value[8:], uc.EncodingBytes)

    return value, nil
}


type TagUnknownType_927C_MakerNote struct {
    MakerNoteType []byte
    MakerNoteBytes []byte
}

func (mn TagUnknownType_927C_MakerNote) String() string {
    parts := make([]string, 20)
    for i, c := range mn.MakerNoteType {
        parts[i] = fmt.Sprintf("%02x", c)
    }

    return fmt.Sprintf("MakerNote<TYPE-ID=[%s]>", strings.Join(parts, " "))
}

func (uc TagUnknownType_927C_MakerNote) ValueBytes() (value []byte, err error) {
    return uc.MakerNoteBytes, nil
}


type TagUnknownType_9101_ComponentsConfiguration struct {
    ConfigurationId int
    ConfigurationBytes []byte
}

func (cc TagUnknownType_9101_ComponentsConfiguration) String() string {
    return fmt.Sprintf("ComponentsConfiguration<ID=[%s] BYTES=%v>", TagUnknownType_9101_ComponentsConfiguration_Names[cc.ConfigurationId], cc.ConfigurationBytes)
}

func (uc TagUnknownType_9101_ComponentsConfiguration) ValueBytes() (value []byte, err error) {
    return uc.ConfigurationBytes, nil
}


// UndefinedValue knows how to resolve the value for most unknown-type tags.
func UndefinedValue(ii IfdIdentity, tagId uint16, valueContext ValueContext, byteOrder binary.ByteOrder) (value interface{}, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    typeLogger.Debugf(nil, "UndefinedValue: IFD=[%v] TAG-ID=(0x%02x)", ii, tagId)

    if ii == ExifIi {
        if tagId == 0x9000 {
            // ExifVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiNoNulValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0xa000 {
            // FlashpixVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiNoNulValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0x9286 {
            // UserComment

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)

            unknownUc := TagUnknownType_9298_UserComment{
                EncodingType: TagUnknownType_9298_UserComment_Encoding_UNDEFINED,
                EncodingBytes: []byte{},
            }

            encoding := valueBytes[:8]
            for encodingIndex, encodingBytes := range TagUnknownType_9298_UserComment_Encodings {
                if bytes.Compare(encoding, encodingBytes) == 0 {
                    // If unknown, return the default rather than what we have
                    // because there will be a big list of NULs (which aren't
                    // functional) and this won't equal the default instance
                    // (above).
                    if encodingIndex == TagUnknownType_9298_UserComment_Encoding_UNDEFINED {
                        return unknownUc, nil
                    } else {
                        uc := TagUnknownType_9298_UserComment{
                            EncodingType: encodingIndex,
                            EncodingBytes: valueBytes[8:],
                        }

                        return uc, nil
                    }
                }
            }

            typeLogger.Warningf(nil, "User-comment encoding not valid. Returning 'unknown' type (the default).")
            return unknownUc, nil
        } else if tagId == 0x927c {
            // MakerNote
// TODO(dustin): !! This is the Wild Wild West. This very well might be a child IFD, but any and all OEM's define their own formats. If we're going to be writing changes and this is complete EXIF (which may not have the first eight bytes), it might be fine. However, if these are just IFDs they'll be relative to the main EXIF, this will invalidate the MakerNote data for IFDs and any other implementations that use offsets unless we can interpret them all. It be best to return to this later and just exclude this from being written for now, though means a loss of a wealth of image metadata.
//                  -> We can also just blindly try to interpret as an IFD and just validate that it's looks good (maybe it will even have a 'next ifd' pointer that we can validate is 0x0).

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)


// TODO(dustin): Doesn't work, but here as an example.
//             ie := NewIfdEnumerate(valueBytes, byteOrder)

// // TODO(dustin): !! Validate types (might have proprietary types, but it might be worth splitting the list between valid and not validate; maybe fail if a certain proportion are invalid, or maybe aren't less then a certain small integer)?
//             ii, err := ie.Collect(0x0)

//             for _, entry := range ii.RootIfd.Entries {
//                 fmt.Printf("ENTRY: 0x%02x %d\n", entry.TagId, entry.TagType)
//             }

            mn := TagUnknownType_927C_MakerNote{
                MakerNoteType: valueBytes[:20],

                // MakerNoteBytes has the whole length of bytes. There's always
                // the chance that the first 20 bytes includes actual data.
                MakerNoteBytes: valueBytes,
            }

            return mn, nil
        } else if tagId == 0x9101 {
            // ComponentsConfiguration

            tt := NewTagType(TypeByte, byteOrder)

            valueBytes, err := tt.ReadByteValues(valueContext)
            log.PanicIf(err)

            for configurationId, configurationBytes := range TagUnknownType_9101_ComponentsConfiguration_Configurations {
                if bytes.Compare(valueBytes, configurationBytes) == 0 {
                    cc := TagUnknownType_9101_ComponentsConfiguration{
                        ConfigurationId: configurationId,
                        ConfigurationBytes: valueBytes,
                    }

                    return cc, nil
                }
            }

            cc := TagUnknownType_9101_ComponentsConfiguration{
                ConfigurationId: TagUnknownType_9101_ComponentsConfiguration_OTHER,
                ConfigurationBytes: valueBytes,
            }

            return cc, nil
        }
    } else if ii == GpsIi {
        if tagId == 0x001c {
            // GPSAreaInformation

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiNoNulValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        } else if tagId == 0x001b {
            // GPSProcessingMethod

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiNoNulValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        }
    } else if ii == ExifIopIi {
        if tagId == 0x0002 {
            // InteropVersion

            tt := NewTagType(TypeAsciiNoNul, byteOrder)

            valueString, err := tt.ReadAsciiNoNulValue(valueContext)
            log.PanicIf(err)

            return TagUnknownType_GeneralString(valueString), nil
        }
    }

// TODO(dustin): !! Still need to do:
//
// complex: 0xa302, 0xa20c, 0x8828
// long: 0xa301, 0xa300

    // 0xa40b is device-specific and unhandled.


    log.Panic(ErrUnhandledUnknownTypedTag)
    return nil, nil
}
