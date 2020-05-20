package exifcommon

import (
	"fmt"
	"strings"

	"github.com/dsoprea/go-logging"
)

// IfdTag describes a single IFD tag and its parent (if any).
type IfdTag struct {
	parentIfdTag *IfdTag
	tagId        uint16
	name         string
}

func NewIfdTag(parentIfdTag *IfdTag, tagId uint16, name string) IfdTag {
	return IfdTag{
		parentIfdTag: parentIfdTag,
		tagId:        tagId,
		name:         name,
	}
}

// ParentIfd returns the IfdTag of this IFD's parent.
func (it IfdTag) ParentIfd() *IfdTag {
	return it.parentIfdTag
}

// TagId returns the tag-ID of this IFD.
func (it IfdTag) TagId() uint16 {
	return it.tagId
}

// Name returns the simple name of this IFD.
func (it IfdTag) Name() string {
	return it.name
}

// String returns a descriptive string.
func (it IfdTag) String() string {
	parentIfdPhrase := ""
	if it.parentIfdTag != nil {
		parentIfdPhrase = fmt.Sprintf(" PARENT=(0x%04x)[%s]", it.parentIfdTag.tagId, it.parentIfdTag.name)
	}

	return fmt.Sprintf("IfdTag<TAG-ID=(0x%04x) NAME=[%s]%s>", it.tagId, it.name, parentIfdPhrase)
}

var (
	// rootStandardIfd is the standard root IFD.
	rootStandardIfd = NewIfdTag(nil, 0x0000, "IFD") // IFD

	// exifStandardIfd is the standard "Exif" IFD.
	exifStandardIfd = NewIfdTag(&rootStandardIfd, 0x8769, "Exif") // IFD/Exif

	// iopStandardIfd is the standard "Iop" IFD.
	iopStandardIfd = NewIfdTag(&exifStandardIfd, 0xA005, "Iop") // IFD/Exif/Iop

	// gpsInfoStandardIfd is the standard "GPS" IFD.
	gpsInfoStandardIfd = NewIfdTag(&rootStandardIfd, 0x8825, "GPSInfo") // IFD/GPSInfo
)

// IfdIdentityPart represents one component in an IFD path.
type IfdIdentityPart struct {
	Name  string
	Index int
}

// String returns a fully-qualified IFD path.
func (iip IfdIdentityPart) String() string {
	if iip.Index > 0 {
		return fmt.Sprintf("%s%d", iip.Name, iip.Index)
	} else {
		return iip.Name
	}
}

// UnindexedString returned a non-fully-qualified IFD path.
func (iip IfdIdentityPart) UnindexedString() string {
	return iip.Name
}

// IfdIdentity represents a single IFD path and provides access to various
// information and representations.
//
// Only global instances can be used for equality checks.
type IfdIdentity struct {
	ifdTag    IfdTag
	parts     []IfdIdentityPart
	ifdPath   string
	fqIfdPath string
}

// NewIfdIdentity returns a new IfdIdentity struct.
func NewIfdIdentity(ifdTag IfdTag, parts ...IfdIdentityPart) (ii *IfdIdentity) {
	ii = &IfdIdentity{
		ifdTag: ifdTag,
		parts:  parts,
	}

	ii.ifdPath = ii.getIfdPath()
	ii.fqIfdPath = ii.getFqIfdPath()

	return ii
}

func (ii *IfdIdentity) getFqIfdPath() string {
	partPhrases := make([]string, len(ii.parts))
	for i, iip := range ii.parts {
		partPhrases[i] = iip.String()
	}

	return strings.Join(partPhrases, "/")
}

func (ii *IfdIdentity) getIfdPath() string {
	partPhrases := make([]string, len(ii.parts))
	for i, iip := range ii.parts {
		partPhrases[i] = iip.UnindexedString()
	}

	return strings.Join(partPhrases, "/")
}

// String returns a fully-qualified IFD path.
func (ii *IfdIdentity) String() string {
	return ii.fqIfdPath
}

// UnindexedString returns a non-fully-qualified IFD path.
func (ii *IfdIdentity) UnindexedString() string {
	return ii.ifdPath
}

// IfdTag returns the tag struct behind this IFD.
func (ii *IfdIdentity) IfdTag() IfdTag {
	return ii.ifdTag
}

// TagId returns the tag-ID of the IFD.
func (ii *IfdIdentity) TagId() uint16 {
	return ii.ifdTag.TagId()
}

// LeafPathPart returns the last right-most path-part, which represents the
// current IFD.
func (ii *IfdIdentity) LeafPathPart() IfdIdentityPart {
	return ii.parts[len(ii.parts)-1]
}

// Name returns the simple name of this IFD.
func (ii *IfdIdentity) Name() string {
	return ii.LeafPathPart().Name
}

// Index returns the index of this IFD (more then one IFD under a parent IFD
// will be numbered [0..n]).
func (ii *IfdIdentity) Index() int {
	return ii.LeafPathPart().Index
}

// Equals returns true if the two IfdIdentity instances are effectively
// identical.
//
// Since there's no way to get a specific fully-qualified IFD path without a
// certain slice of parts and all other fields are also derived from this,
// checking that the fully-qualified IFD path is equals is sufficient.
func (ii *IfdIdentity) Equals(ii2 *IfdIdentity) bool {
	return ii.String() == ii2.String()
}

// NewChild creates an IfdIdentity for an IFD that is a child of the current
// IFD.
func (ii *IfdIdentity) NewChild(childIfdTag IfdTag, index int) (iiChild *IfdIdentity) {
	if *childIfdTag.parentIfdTag != ii.ifdTag {
		log.Panicf("can not add child; we are not the parent:\nUS=%v\nCHILD=%v", ii.ifdTag, childIfdTag)
	}

	childPart := IfdIdentityPart{childIfdTag.name, index}
	childParts := append(ii.parts, childPart)

	iiChild = NewIfdIdentity(childIfdTag, childParts...)
	return iiChild
}

// NewSibling creates an IfdIdentity for an IFD that is a sibling to the current
// one.
func (ii *IfdIdentity) NewSibling(index int) (iiSibling *IfdIdentity) {
	parts := make([]IfdIdentityPart, len(ii.parts))

	copy(parts, ii.parts)
	parts[len(parts)-1].Index = index

	iiSibling = NewIfdIdentity(ii.ifdTag, parts...)
	return iiSibling
}

var (
	// IfdStandardIfdIdentity represents the IFD path for IFD0.
	IfdStandardIfdIdentity = NewIfdIdentity(rootStandardIfd, IfdIdentityPart{"IFD", 0})

	// IfdExifStandardIfdIdentity represents the IFD path for IFD0/Exif0.
	IfdExifStandardIfdIdentity = IfdStandardIfdIdentity.NewChild(exifStandardIfd, 0)

	// IfdExifIopStandardIfdIdentity represents the IFD path for IFD0/Exif0/Iop0.
	IfdExifIopStandardIfdIdentity = IfdExifStandardIfdIdentity.NewChild(iopStandardIfd, 0)

	// IfdGPSInfoStandardIfdIdentity represents the IFD path for IFD0/GPSInfo0.
	IfdGpsInfoStandardIfdIdentity = IfdStandardIfdIdentity.NewChild(gpsInfoStandardIfd, 0)

	// Ifd1StandardIfdIdentity represents the IFD path for IFD1.
	Ifd1StandardIfdIdentity = NewIfdIdentity(rootStandardIfd, IfdIdentityPart{"IFD", 1})
)

var (
	// RELEASE(dustin): These are for backwards-compatibility. These used to be strings but are now IfdIdentity structs and the newer "StandardIfdIdentity" symbols above should be used instead. These will be removed in the next release.

	IfdPathStandard        = IfdStandardIfdIdentity
	IfdPathStandardExif    = IfdExifStandardIfdIdentity
	IfdPathStandardExifIop = IfdExifIopStandardIfdIdentity
	IfdPathStandardGps     = IfdGpsInfoStandardIfdIdentity
)
