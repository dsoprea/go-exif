package exif

import (
    "testing"

    "github.com/dsoprea/go-logging"
)

var (
    invalidIi = IfdIdentity{
        ParentIfdName: "invalid-parent",
        IfdName: "invalid-child",
    }
)

func TestGet(t *testing.T) {
    ti := NewTagIndex()

    it, err := ti.Get(RootIi, 0x10f)
    log.PanicIf(err)

    if it.Is("IFD", 0x10f) == false || it.IsName("IFD", "Make") == false {
        t.Fatalf("tag info not correct")
    }
}

func TestGetWithName(t *testing.T) {
    ti := NewTagIndex()

    it, err := ti.GetWithName(RootIi, "Make")
    log.PanicIf(err)

    if it.Is("IFD", 0x10f) == false || it.Is("IFD", 0x10f) == false {
        t.Fatalf("tag info not correct")
    }
}

func TestIfdTagNameWithIdOrFail_Miss(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            if err.Error() != "tag-ID (0x1234) under parent IFD [Exif] not associated with a child IFD" {
                log.Panic(err)
            }
        }
    }()

    IfdTagNameWithIdOrFail(IfdExif, 0x1234)

    t.Fatalf("expected failing for invalid IFD tag-ID")
}

func TestIfdTagNameWithIdOrFail_RootChildIfd(t *testing.T) {
    name := IfdTagNameWithIdOrFail(IfdStandard, IfdExifId)
    if name != IfdExif {
        t.Fatalf("expected EXIF IFD name for hit on EXIF IFD")
    }
}

func TestIfdTagNameWithIdOrFail_ChildChildIfd(t *testing.T) {
    name := IfdTagNameWithIdOrFail(IfdExif, IfdIopId)
    if name != IfdIop {
        t.Fatalf("expected IOP IFD name for hit on IOP IFD")
    }
}

func TestIfdTagNameWithId_Hit(t *testing.T) {
    name, found := IfdTagNameWithId(IfdExif, IfdIopId)
    if found != true {
        t.Fatalf("could not get name for IOP IFD tag-ID")
    } else if name != IfdIop {
        t.Fatalf("expected IOP IFD name for hit on IOP IFD")
    }
}

func TestIfdTagNameWithId_Miss(t *testing.T) {
    name, found := IfdTagNameWithId(IfdExif, 0x1234)
    if found != false {
        t.Fatalf("expected failure for invalid IFD iag-ID under EXIF IFD")
    } else if name != "" {
        t.Fatalf("expected empty IFD name for miss")
    }
}

func TestIfdTagIdWithIdentity_Hit(t *testing.T) {
    tagId, found := IfdTagIdWithIdentity(GpsIi)
    if found != true {
        t.Fatalf("could not get tag-ID for II")
    } else if tagId != IfdGpsId {
        t.Fatalf("incorrect tag-ID returned for II")
    }
}

func TestIfdTagIdWithIdentity_Miss(t *testing.T) {
    tagId, found := IfdTagIdWithIdentity(invalidIi)
    if found != false {
        t.Fatalf("expected failure")
    } else if tagId != 0 {
        t.Fatalf("expected tag-ID of (0) for failure")
    }
}

func TestIfdTagIdWithIdentityOrFail_Hit(t *testing.T) {
    IfdTagIdWithIdentityOrFail(GpsIi)
}

func TestIfdTagIdWithIdentityOrFail_Miss(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            if err.Error() != "no tag for invalid IFD identity: IfdIdentity<PARENT-NAME=[invalid-parent] NAME=[invalid-child]>" {
                log.Panic(err)
            }
        }
    }()

    IfdTagIdWithIdentityOrFail(invalidIi)

    t.Fatalf("invalid II didn't panic")
}

func TestIfdIdWithIdentity_Hit(t *testing.T) {
    id := IfdIdWithIdentity(GpsIi)
    if id != 3 {
        t.Fatalf("II doesn't have the right ID")
    }
}

func TestIfdIdWithIdentity_Miss(t *testing.T) {
    id := IfdIdWithIdentity(invalidIi)
    if id != 0 {
        t.Fatalf("II doesn't have the right ID for a miss")
    }
}

func TestIfdIdWithIdentityOrFail_Hit(t *testing.T) {
    id := IfdIdWithIdentityOrFail(GpsIi)
    if id != 3 {
        t.Fatalf("II doesn't have the right ID")
    }
}

func TestIfdIdWithIdentityOrFail_Miss(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            if err.Error() != "IFD not valid: IfdIdentity<PARENT-NAME=[invalid-parent] NAME=[invalid-child]>" {
                log.Panic(err)
            }
        }
    }()

    IfdIdWithIdentityOrFail(invalidIi)

    t.Fatalf("invalid II doesn't panic")
}

func TestIfdIdOrFail_Hit(t *testing.T) {
    ii, id := IfdIdOrFail(IfdStandard, IfdExif)
    if ii != ExifIi {
        t.Fatalf("wrong II for IFD returned")
    } else if id != 2 {
        t.Fatalf("wrong ID for II returned")
    }
}

func TestIfdIdOrFail_Miss(t *testing.T) {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            if err.Error() != "IFD is not valid: [IFD] [invalid-ifd]" {
                log.Panic(err)
            }
        }
    }()

    IfdIdOrFail(IfdStandard, "invalid-ifd")

    t.Fatalf("expected panic for invalid IFD")
}

func TestIfdId_Hit(t *testing.T) {
    ii, id := IfdId(IfdStandard, IfdExif)
    if ii != ExifIi {
        t.Fatalf("wrong II for IFD returned")
    } else if id != 2 {
        t.Fatalf("wrong ID for II returned")
    }
}

func TestIfdId_Miss(t *testing.T) {
    ii, id := IfdId(IfdStandard, "invalid-ifd")
    if id != 0 {
        t.Fatalf("non-zero ID returned for invalid IFD")
    } else if ii != ZeroIi {
        t.Fatalf("expected zero-instance result for miss")
    }
}
