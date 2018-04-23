package exif

import (
    "os"
    "path"
    "fmt"
    "errors"

    "gopkg.in/yaml.v2"
    "github.com/dsoprea/go-logging"
)

const (
    IfdStandard = "IFD"
    IfdExif = "Exif"
    IfdGps = "GPSInfo"
    IfdIop = "Iop"
)

var (
    tagDataFilepath = ""

    validIfds = []string {
        IfdStandard,
        IfdExif,
        IfdGps,
        IfdIop,
    }

    IfdTagIds = map[string]uint16 {
        IfdExif: 0x8769,
        IfdGps: 0x8825,
        IfdIop: 0xA005,
    }

    // IfdTagNames is populated in the init(), below.
    IfdTagNames = map[uint16]string {}
)

var (
    tagsLogger = log.NewLogger("exif.tags")
    ErrTagNotFound = errors.New("tag not found")
)


// File structures.

type encodedTag struct {
    // id is signed, here, because YAML doesn't have enough information to
    // support unsigned.
    Id int `yaml:"id"`
    Name string `yaml:"name"`
    TypeName string `yaml:"type_name"`
}


// Indexing structures.

type IndexedTag struct {
    Id uint16
    Name string
    Ifd string
    Type uint16
}

func (it IndexedTag) String() string {
    return fmt.Sprintf("TAG<ID=(0x%04x) NAME=[%s] IFD=[%s]>", it.Id, it.Name, it.Ifd)
}

func (it IndexedTag) IsName(ifd, name string) bool {
    return it.Name == name && it.Ifd == ifd
}

func (it IndexedTag) Is(id uint16) bool {
    return it.Id == id
}

type TagIndex struct {
    tagsByIfd map[string]map[uint16]*IndexedTag
}

func NewTagIndex() *TagIndex {
    ti := new(TagIndex)

    err := ti.load()
    log.PanicIf(err)

    return ti
}

func (ti *TagIndex) load() (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()


    // Read static data.

    f, err := os.Open(tagDataFilepath)
    log.PanicIf(err)

    d := yaml.NewDecoder(f)

    encodedIfds := make(map[string][]encodedTag)

    err = d.Decode(encodedIfds)
    log.PanicIf(err)


    // Load structure.

    tagsByIfd := make(map[string]map[uint16]*IndexedTag)

    count := 0
    for ifdName, tags := range encodedIfds {
        for _, tagInfo := range tags {
            tagId := uint16(tagInfo.Id)
            tagName := tagInfo.Name
            tagTypeName := tagInfo.TypeName

// TODO(dustin): !! Non-standard types but present types. Ignore for right now.
if tagTypeName == "SSHORT" || tagTypeName == "FLOAT" || tagTypeName == "DOUBLE" {
    continue
}

            tagTypeId, found := TypeNamesR[tagTypeName]
            if found == false {
                log.Panicf("type [%s] for [%s] not valid", tagTypeName, tagName)
                continue
            }

            tag := &IndexedTag{
                Ifd: ifdName,
                Id: tagId,
                Name: tagName,
                Type: tagTypeId,
            }

            family, found := tagsByIfd[ifdName]
            if found == false {
                family = make(map[uint16]*IndexedTag)
                tagsByIfd[ifdName] = family
            }

            if _, found := family[tagId]; found == true {
                log.Panicf("tag-ID defined more than once for IFD [%s]: (%02x)", ifdName, tagId)
            }

            family[tagId] = tag

            count++
        }
    }

    ti.tagsByIfd = tagsByIfd

    tagsLogger.Debugf(nil, "(%d) tags loaded.", count)

    return nil
}

func (ti *TagIndex) Get(ifdName string, id uint16) (it *IndexedTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    family, found := ti.tagsByIfd[ifdName]
    if found == false {
        log.Panic(ErrTagNotFound)
    }

    it, found = family[id]
    if found == false {
        log.Panic(ErrTagNotFound)
    }

    return it, nil
}

// IfdName returns the known index name for the tags that are expected/allowed
// for the IFD. If there's an error, returns "". If returns "", the IFD should
// be skipped.
func IfdName(ifdName string, ifdIndex int) string {
    // There's an IFD0 and IFD1, but the others must be unique.
    if ifdName == IfdStandard && ifdIndex > 1 {
        tagsLogger.Errorf(nil, "The 'IFD' IFD can not occur more than twice: [%s]. Ignoring IFD.", ifdName)
        return ""
    } else if ifdName != IfdStandard && ifdIndex > 0 {
        tagsLogger.Errorf(nil, "Only the 'IFD' IFD can be repeated: [%s]. Ignoring IFD.", ifdName)
        return ""
    }

    return ifdName
}

// IsIfdTag returns true if the given tag points to a child IFD block.
func IsIfdTag(tagId uint16) (name string, found bool) {
    name, found = IfdTagNames[tagId]
    return name, found
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath := path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
    tagDataFilepath = path.Join(assetsPath, "tags.yaml")

    for name, tagId := range IfdTagIds {
        IfdTagNames[tagId] = name
    }
}
