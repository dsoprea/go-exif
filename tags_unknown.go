package exif

import (
    "fmt"
    "strings"

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
