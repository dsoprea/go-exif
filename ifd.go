package exif

import (
	"fmt"
	"strings"

	"github.com/dsoprea/go-logging"
)

const (
	// The root IFD types (ifd0, ifd1).
	IfdStandard = "IFD"

	// Child IFD types.
	IfdExif = "Exif"
	IfdGps  = "GPSInfo"
	IfdIop  = "Iop"

	// Tag IDs for child IFDs.
	IfdExifId = 0x8769
	IfdGpsId  = 0x8825
	IfdIopId  = 0xA005

	// Just a placeholder.
	IfdRootId = 0x0000
)

type IfdNameAndIndex struct {
	Ii    IfdIdentity
	Index int
}

var (
	// TODO(dustin): !! Get rid of this in favor of one of the two lookups, just below.
	validIfds = []string{
		IfdStandard,
		IfdExif,
		IfdGps,
		IfdIop,
	}

	// A lookup for IFDs by their parents.
	// TODO(dustin): !! We should switch to indexing by their unique integer IDs (defined below) rather than exposing ourselves to non-unique IFD names (even if we *do* manage the naming ourselves).
	IfdTagIds = map[string]map[string]uint16{
		"": map[string]uint16{
			// A root IFD type. Not allowed to be a child (tag-based) IFD.
			IfdStandard: 0x0,
		},

		IfdStandard: map[string]uint16{
			IfdExif: IfdExifId,
			IfdGps:  IfdGpsId,
		},

		IfdExif: map[string]uint16{
			IfdIop: IfdIopId,
		},
	}

	// IfdTagNames contains the tag ID-to-name mappings and is populated by
	// init().
	IfdTagNames = map[string]map[uint16]string{}

	// IFD Identities. These are often how we refer to IFDs, from call to call.

	// The NULL-type instance for search misses and empty responses.
	ZeroIi = IfdIdentity{}

	RootIi    = IfdIdentity{IfdName: IfdStandard}
	ExifIi    = IfdIdentity{ParentIfdName: IfdStandard, IfdName: IfdExif}
	GpsIi     = IfdIdentity{ParentIfdName: IfdStandard, IfdName: IfdGps}
	ExifIopIi = IfdIdentity{ParentIfdName: IfdExif, IfdName: IfdIop}

	// Produce a list of unique IDs for each IFD that we can pass around (so we
	// don't always have to be comparing parent and child names).
	//
	// For lack of need, this is just static.
	//
	// (0) is reserved for not-found/miss responses.
	IfdIds = map[IfdIdentity]int{
		RootIi:    1,
		ExifIi:    2,
		GpsIi:     3,
		ExifIopIi: 4,
	}

	IfdDesignations = map[string]IfdNameAndIndex{
		"ifd0": {RootIi, 0},
		"ifd1": {RootIi, 1},
		"exif": {ExifIi, 0},
		"gps":  {GpsIi, 0},
		"iop":  {ExifIopIi, 0},
	}

	IfdDesignationsR = make(map[IfdNameAndIndex]string)
)

var (
	ifdLogger = log.NewLogger("exif.ifd")
)

func IfdDesignation(ii IfdIdentity, index int) string {
	if ii == RootIi {
		return fmt.Sprintf("%s%d", ii.IfdName, index)
	} else {
		return ii.IfdName
	}
}

type IfdIdentity struct {
	ParentIfdName string
	IfdName       string
}

func (ii IfdIdentity) String() string {
	return fmt.Sprintf("IfdIdentity<PARENT-NAME=[%s] NAME=[%s]>", ii.ParentIfdName, ii.IfdName)
}

func (ii IfdIdentity) Id() int {
	return IfdIdWithIdentityOrFail(ii)
}

type MappedIfd struct {
	ParentTagId uint16
	Path        []string

	Name     string
	TagId    uint16
	Children map[uint16]*MappedIfd
}

func (mi *MappedIfd) String() string {
	pathPhrase := mi.PathPhrase()
	return fmt.Sprintf("MappedIfd<(0x%04X) [%s] PATH=[%s]>", mi.TagId, mi.Name, pathPhrase)
}

func (mi *MappedIfd) PathPhrase() string {
	return strings.Join(mi.Path, "/")
}

// IfdMapping describes all of the IFDs that we currently recognize.
type IfdMapping struct {
	rootNode *MappedIfd
}

func NewIfdMapping() (ifdMapping *IfdMapping) {
	rootNode := &MappedIfd{
		Path:     make([]string, 0),
		Children: make(map[uint16]*MappedIfd),
	}

	return &IfdMapping{
		rootNode: rootNode,
	}
}

func (im *IfdMapping) Get(parentPlacement []uint16) (childIfd *MappedIfd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ptr := im.rootNode
	for _, tagId := range parentPlacement {
		if descendantPtr, found := ptr.Children[tagId]; found == false {
			log.Panicf(fmt.Sprintf("ifd child with tag-ID (%04x) not registered", tagId))
		} else {
			ptr = descendantPtr
		}
	}

	return ptr, nil
}

func (im *IfdMapping) GetWithPath(pathPhrase string) (mi *MappedIfd, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	path := strings.Split(pathPhrase, "/")
	ptr := im.rootNode

	for _, name := range path {
		var hit *MappedIfd
		for _, mi := range ptr.Children {
			if mi.Name == name {
				hit = mi
				break
			}
		}

		if hit == nil {
			log.Panicf(fmt.Sprintf("ifd child with name [%s] not registered", name))
		}

		ptr = hit
	}

	return ptr, nil
}

// Add puts the given IFD at the given position of the tree. The position of the
// tree is referred to as the placement and is represented by a set of tag-IDs,
// where the leftmost is the root tag and the tags going to the right are
// progressive descendants.
func (im *IfdMapping) Add(parentPlacement []uint16, tagId uint16, name string) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	ptr, err := im.Get(parentPlacement)
	log.PanicIf(err)

	path := make([]string, len(parentPlacement)+1)
	if len(parentPlacement) > 0 {
		copy(path, ptr.Path)
	}

	path[len(path)-1] = name

	childIfd := &MappedIfd{
		ParentTagId: ptr.TagId,
		Path:        path,
		Name:        name,
		TagId:       tagId,
		Children:    make(map[uint16]*MappedIfd),
	}

	if _, found := ptr.Children[tagId]; found == true {
		log.Panicf("child IFD with tag-ID (%04x) already registered under IFD [%s] with tag-ID (%04x)", tagId, ptr.Name, ptr.TagId)
	}

	ptr.Children[tagId] = childIfd

	return nil
}

func (im *IfdMapping) dumpLineages(stack []*MappedIfd, input []string) (output []string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	currentIfd := stack[len(stack)-1]

	output = input
	for _, childIfd := range currentIfd.Children {
		stackCopy := make([]*MappedIfd, len(stack)+1)

		copy(stackCopy, stack)
		stackCopy[len(stack)] = childIfd

		// Add to output, but don't include the obligatory root node.
		parts := make([]string, len(stackCopy)-1)
		for i, mi := range stackCopy[1:] {
			parts[i] = mi.Name
		}

		output = append(output, strings.Join(parts, "/"))

		output, err = im.dumpLineages(stackCopy, output)
		log.PanicIf(err)
	}

	return output, nil
}

func (im *IfdMapping) DumpLineages() (output []string, err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	stack := []*MappedIfd{im.rootNode}
	output = make([]string, 0)

	output, err = im.dumpLineages(stack, output)
	log.PanicIf(err)

	return output, nil
}

func LoadStandardIfds(im *IfdMapping) (err error) {
	defer func() {
		if state := recover(); state != nil {
			err = log.Wrap(state.(error))
		}
	}()

	err = im.Add([]uint16{}, IfdRootId, IfdStandard)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId}, IfdExifId, IfdExif)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId, IfdExifId}, IfdIopId, IfdIop)
	log.PanicIf(err)

	err = im.Add([]uint16{IfdRootId}, IfdGpsId, IfdGps)
	log.PanicIf(err)

	return nil
}

func init() {
	for ifdName, tags := range IfdTagIds {
		tagsR := make(map[uint16]string)

		for tagName, tagId := range tags {
			tagsR[tagId] = tagName
		}

		IfdTagNames[ifdName] = tagsR
	}

	for designation, ni := range IfdDesignations {
		IfdDesignationsR[ni] = designation
	}
}
