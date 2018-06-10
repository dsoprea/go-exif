package exif

import (
    "fmt"

    "github.com/dsoprea/go-logging"
)

const (
    // The root IFD types (ifd0, ifd1).
    IfdStandard = "IFD"

    // Child IFD types.
    IfdExif = "Exif"
    IfdGps = "GPSInfo"
    IfdIop = "Iop"

    // Tag IDs for child IFDs.
    IfdExifId = 0x8769
    IfdGpsId = 0x8825
    IfdIopId = 0xA005
)

type IfdNameAndIndex struct {
    Ii IfdIdentity
    Index int
}

var (
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

    IfdDesignations = map[string]IfdNameAndIndex {
        "ifd0": { RootIi, 0 },
        "ifd1": { RootIi, 1 },
        "exif": { ExifIi, 0 },
        "gps": { GpsIi, 0 },
        "iop": { ExifIopIi, 0 },
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
    IfdName string
}

func (ii IfdIdentity) String() string {
    return fmt.Sprintf("IfdIdentity<PARENT-NAME=[%s] NAME=[%s]>", ii.ParentIfdName, ii.IfdName)
}

func (ii IfdIdentity) Id() int {
    return IfdIdWithIdentityOrFail(ii)
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
