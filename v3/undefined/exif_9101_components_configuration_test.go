package exifundefined

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestTagExif9101ComponentsConfiguration_String(t *testing.T) {
	ut := TagExif9101ComponentsConfiguration{
		ConfigurationId:    TagUndefinedType_9101_ComponentsConfiguration_RGB,
		ConfigurationBytes: []byte{0x11, 0x22, 0x33, 0x44},
	}

	s := ut.String()

	if s != "Exif9101ComponentsConfiguration<ID=[RGB] BYTES=[17 34 51 68]>" {
		t.Fatalf("String not correct: [%s]", s)
	}
}

func TestCodecExif9101ComponentsConfiguration_Encode(t *testing.T) {
	configurationBytes := []byte(TagUndefinedType_9101_ComponentsConfiguration_Names[TagUndefinedType_9101_ComponentsConfiguration_RGB])

	ut := TagExif9101ComponentsConfiguration{
		ConfigurationId:    TagUndefinedType_9101_ComponentsConfiguration_RGB,
		ConfigurationBytes: configurationBytes,
	}

	codec := CodecExif9101ComponentsConfiguration{}

	encoded, unitCount, err := codec.Encode(ut, exifcommon.TestDefaultByteOrder)
	log.PanicIf(err)

	if bytes.Equal(encoded, configurationBytes) != true {
		exifcommon.DumpBytesClause(encoded)

		t.Fatalf("Encoded bytes not correct: %v", encoded)
	} else if unitCount != uint32(len(configurationBytes)) {
		t.Fatalf("Unit-count not correct: (%d)", unitCount)
	}

	s := string(configurationBytes)

	if s != TagUndefinedType_9101_ComponentsConfiguration_Names[TagUndefinedType_9101_ComponentsConfiguration_RGB] {
		t.Fatalf("Recovered configuration name not correct: [%s]", s)
	}
}

func TestCodecExif9101ComponentsConfiguration_Decode(t *testing.T) {
	configurationBytes := TagUndefinedType_9101_ComponentsConfiguration_Configurations[TagUndefinedType_9101_ComponentsConfiguration_RGB]

	ut := TagExif9101ComponentsConfiguration{
		ConfigurationId:    TagUndefinedType_9101_ComponentsConfiguration_RGB,
		ConfigurationBytes: configurationBytes,
	}

	rawValueOffset := configurationBytes

	valueContext := exifcommon.NewValueContext(
		"",
		0,
		uint32(len(configurationBytes)),
		0,
		rawValueOffset,
		nil,
		exifcommon.TypeUndefined,
		exifcommon.TestDefaultByteOrder)

	codec := CodecExif9101ComponentsConfiguration{}

	value, err := codec.Decode(valueContext)
	log.PanicIf(err)

	if reflect.DeepEqual(value, ut) != true {
		t.Fatalf("Decoded value not correct: %s != %s\n", value, ut)
	}
}
