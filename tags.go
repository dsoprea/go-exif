package exif

import (
    "os"
    "path"
    "fmt"

    "gopkg.in/yaml.v2"
    "github.com/dsoprea/go-logging"
)

const (
    IfdStandard = "IFD"

    IfdExif = "Exif"
    IfdGps = "GPSInfo"
    IfdIop = "Iop"

    IfdExifId = 0x8769
    IfdGpsId = 0x8825
    IfdIopId = 0xA005

    ThumbnailOffsetTagId = 0x0201
    ThumbnailSizeTagId = 0x0202
)

var (
    tagDataFilepath = ""

// TODO(dustin): !! Get rid of this in favor of one of the two lookups, just below.
    validIfds = []string {
        IfdStandard,
        IfdExif,
        IfdGps,
        IfdIop,
    }

    // A lookup for IFDs by their parents.
// TODO(dustin): !! We should switch to indexing by their unique integer IDs (defined below) rather than exposing ourselves to non-unique IFD names (even if we *do* manage the naming ourselves).
    IfdTagIds = map[string]map[string]uint16 {
        "": map[string]uint16 {
            // A root IFD type. Not allowed to be a child (tag-based) IFD.
            IfdStandard: 0x0,
        },

        IfdStandard: map[string]uint16 {
            IfdExif: IfdExifId,
            IfdGps: IfdGpsId,
        },

        IfdExif: map[string]uint16 {
            IfdIop: IfdIopId,
        },
    }

    // IfdTagNames contains the tag ID-to-name mappings and is populated by
    // init().
    IfdTagNames = map[string]map[uint16]string {}

    // IFD Identities. These are often how we refer to IFDs, from call to call.

    // The NULL-type instance for search misses and empty responses.
    ZeroIi = IfdIdentity{}

    RootIi = IfdIdentity{ IfdName: IfdStandard }
    ExifIi = IfdIdentity{ ParentIfdName: IfdStandard, IfdName: IfdExif }
    GpsIi = IfdIdentity{ ParentIfdName: IfdStandard, IfdName: IfdGps }
    ExifIopIi = IfdIdentity{ ParentIfdName: IfdExif, IfdName: IfdIop }

    // Produce a list of unique IDs for each IFD that we can pass around (so we
    // don't always have to be comparing parent and child names).
    //
    // For lack of need, this is just static.
    //
    // (0) is reserved for not-found/miss responses.
    IfdIds = map[IfdIdentity]int {
        RootIi: 1,
        ExifIi: 2,
        GpsIi: 3,
        ExifIopIi: 4,
    }

    // tagsWithoutAlignment is a tag-lookup for tags whose value size won't
    // necessarily be a multiple of its tag-type.
    tagsWithoutAlignment = map[uint16]struct{} {
        // The thumbnail offset is stored as a long, but its data is a binary
        // blob (not a slice of longs).
        ThumbnailOffsetTagId: struct{}{},
    }

    tagIndex *TagIndex
)

var (
    tagsLogger = log.NewLogger("exif.tags")
)


type IfdIdentity struct {
    ParentIfdName string
    IfdName string
}

func (ii IfdIdentity) String() string {
    return fmt.Sprintf("IfdIdentity<PARENT-NAME=[%s] NAME=[%s]>", ii.ParentIfdName, ii.IfdName)
}

func (ii IfdIdentity) Id() int {
    return IfdIdWithIdentityOrFail(ii)
}


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

func (it IndexedTag) Is(ifd string, id uint16) bool {
    return it.Id == id && it.Ifd == ifd
}

type TagIndex struct {
    tagsByIfd map[string]map[uint16]*IndexedTag
    tagsByIfdR map[string]map[string]*IndexedTag
}

func NewTagIndex() *TagIndex {
    ti := new(TagIndex)
    return ti
}

func (ti *TagIndex) load() (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    // Already loaded.
    if ti.tagsByIfd != nil {
        return nil
    }


    // Read static data.

    encodedIfds := make(map[string][]encodedTag)

    err = yaml.Unmarshal([]byte(tagsYaml), encodedIfds)
    log.PanicIf(err)


    // Load structure.

    tagsByIfd := make(map[string]map[uint16]*IndexedTag)
    tagsByIfdR := make(map[string]map[string]*IndexedTag)

    count := 0
    for ifdName, tags := range encodedIfds {
        for _, tagInfo := range tags {
            tagId := uint16(tagInfo.Id)
            tagName := tagInfo.Name
            tagTypeName := tagInfo.TypeName

// TODO(dustin): !! Non-standard types, but found in real data. Ignore for right now.
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


            // Store by ID.

            family, found := tagsByIfd[ifdName]
            if found == false {
                family = make(map[uint16]*IndexedTag)
                tagsByIfd[ifdName] = family
            }

            if _, found := family[tagId]; found == true {
                log.Panicf("tag-ID defined more than once for IFD [%s]: (%02x)", ifdName, tagId)
            }

            family[tagId] = tag


            // Store by name.

            familyR, found := tagsByIfdR[ifdName]
            if found == false {
                familyR = make(map[string]*IndexedTag)
                tagsByIfdR[ifdName] = familyR
            }

            if _, found := familyR[tagName]; found == true {
                log.Panicf("tag-name defined more than once for IFD [%s]: (%s)", ifdName, tagName)
            }

            familyR[tagName] = tag

            count++
        }
    }

    ti.tagsByIfd = tagsByIfd
    ti.tagsByIfdR = tagsByIfdR

    tagsLogger.Debugf(nil, "(%d) tags loaded.", count)

    return nil
}

// Get returns information about the non-IFD tag.
func (ti *TagIndex) Get(ii IfdIdentity, id uint16) (it *IndexedTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ti.load()
    log.PanicIf(err)

    family, found := ti.tagsByIfd[ii.IfdName]
    if found == false {
        log.Panic(ErrTagNotFound)
    }

    it, found = family[id]
    if found == false {
        log.Panic(ErrTagNotFound)
    }

    return it, nil
}

// Get returns information about the non-IFD tag.
func (ti *TagIndex) GetWithName(ii IfdIdentity, name string) (it *IndexedTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    err = ti.load()
    log.PanicIf(err)

    it, found := ti.tagsByIfdR[ii.IfdName][name]
    if found != true {
        log.Panic(ErrTagNotFound)
    }

    return it, nil
}


// IfdTagWithId returns true if the given tag points to a child IFD block.
func IfdTagNameWithIdOrFail(parentIfdName string, tagId uint16) string {
    name, found := IfdTagNameWithId(parentIfdName, tagId)
    if found == false {
        log.Panicf("tag-ID (0x%02x) under parent IFD [%s] not associated with a child IFD", tagId, parentIfdName)
    }

    return name
}

// IfdTagWithId returns true if the given tag points to a child IFD block.

// TODO(dustin): !! Rewrite to take an IfdIdentity, instead. We shouldn't expect that IFD names are globally unique.

func IfdTagNameWithId(parentIfdName string, tagId uint16) (name string, found bool) {
    if tags, found := IfdTagNames[parentIfdName]; found == true {
        if name, found = tags[tagId]; found == true {
            return name, true
        }
    }

    return "", false
}

// IfdTagIdWithIdentity returns true if the given tag points to a child IFD
// block.
func IfdTagIdWithIdentity(ii IfdIdentity) (tagId uint16, found bool) {
    if tags, found := IfdTagIds[ii.ParentIfdName]; found == true {
        if tagId, found = tags[ii.IfdName]; found == true {
            return tagId, true
        }
    }

    return 0, false
}

func IfdTagIdWithIdentityOrFail(ii IfdIdentity) (tagId uint16) {
    if tags, found := IfdTagIds[ii.ParentIfdName]; found == true {
        if tagId, found = tags[ii.IfdName]; found == true {
            if tagId == 0 {
                // This IFD is not the type that can be linked to from a tag.
                log.Panicf("not a child IFD: [%s]", ii.IfdName)
            }

            return tagId
        }
    }

    log.Panicf("no tag for invalid IFD identity: %v", ii)
    return 0
}

func IfdIdWithIdentity(ii IfdIdentity) int {
    id, _ := IfdIds[ii]
    return id
}

func IfdIdWithIdentityOrFail(ii IfdIdentity) int {
    id, _ := IfdIds[ii]
    if id == 0 {
        log.Panicf("IFD not valid: %v", ii)
    }

    return id
}

func IfdIdOrFail(parentIfdName, ifdName string) (ii IfdIdentity, id int) {
    ii, id = IfdId(parentIfdName, ifdName)
    if id == 0 {
        log.Panicf("IFD is not valid: [%s] [%s]", parentIfdName, ifdName)
    }

    return ii, id
}

func IfdId(parentIfdName, ifdName string) (ii IfdIdentity, id int) {
    ii = IfdIdentity{
        ParentIfdName: parentIfdName,
        IfdName: ifdName,
    }

    id, found := IfdIds[ii]
    if found != true {
        return IfdIdentity{}, 0
    }

    return ii, id
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath := path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
    tagDataFilepath = path.Join(assetsPath, "tags.yaml")

    for ifdName, tags := range IfdTagIds {
        tagsR := make(map[uint16]string)

        for tagName, tagId := range tags {
            tagsR[tagId] = tagName
        }

        IfdTagNames[ifdName] = tagsR
    }

    tagIndex = NewTagIndex()
}
