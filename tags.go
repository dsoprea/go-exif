package exif

import (
    "os"
    "path"
    "fmt"
    "errors"

    "gopkg.in/yaml.v2"
    "github.com/dsoprea/go-logging"
)

var (
    tagDataFilepath = ""
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
}


// Indexing structures.

type IndexedTag struct {
    Id uint16
    Name string
    Ifd string
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
    tagsById map[uint16]*IndexedTag
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

    tagsById := make(map[uint16]*IndexedTag)
    tagsByIfd := make(map[string]map[uint16]*IndexedTag)

    count := 0
    for ifdName, tags := range encodedIfds {
        for _, tagInfo := range tags {
            tagId := uint16(tagInfo.Id)
            tagName := tagInfo.Name

            tag := &IndexedTag{
                Ifd: ifdName,
                Id: tagId,
                Name: tagName,
            }

            if _, found := tagsById[tagId]; found == true {
                log.Panicf("tag-ID defined more than once: (%02x)", tagId)
            }

            tagsById[tagId] = tag

            family, found := tagsByIfd[ifdName]
            if found == false {
                family = make(map[uint16]*IndexedTag)
                tagsByIfd[ifdName] = family
            }

            family[tagId] = tag

            count++
        }
    }

    ti.tagsById = tagsById
    ti.tagsByIfd = tagsByIfd

    tagsLogger.Debugf(nil, "(%d) tags loaded.", count)

    return nil
}

func (ti *TagIndex) GetWithTagId(id uint16) (it *IndexedTag, err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    it, found := ti.tagsById[id]
    if found == false {
        log.Panic(ErrTagNotFound)
    }

    return it, nil
}

func init() {
    goPath := os.Getenv("GOPATH")
    if goPath == "" {
        log.Panicf("GOPATH is empty")
    }

    assetsPath := path.Join(goPath, "src", "github.com", "dsoprea", "go-exif", "assets")
    tagDataFilepath = path.Join(assetsPath, "tags.yaml")
}
