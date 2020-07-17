package main

import (
	"bytes"
	"fmt"
	"path"
	"reflect"
	"testing"

	"encoding/json"
	"io/ioutil"
	"os/exec"

	"github.com/dsoprea/go-logging"

	"github.com/dsoprea/go-exif/v3/common"
)

func TestMain(t *testing.T) {
	appFilepath := getAppFilepath()
	testImageFilepath := getTestImageFilepath()

	cmd := exec.Command(
		"go", "run", appFilepath,
		"--filepath", testImageFilepath)

	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = b

	err := cmd.Run()
	actual := b.String()

	if err != nil {
		fmt.Printf(actual)
		log.Panic(err)
	}

	expected := `IFD-PATH=[IFD] ID=(0x010f) NAME=[Make] COUNT=(6) TYPE=[ASCII] VALUE=[Canon]
IFD-PATH=[IFD] ID=(0x0110) NAME=[Model] COUNT=(22) TYPE=[ASCII] VALUE=[Canon EOS 5D Mark III]
IFD-PATH=[IFD] ID=(0x0112) NAME=[Orientation] COUNT=(1) TYPE=[SHORT] VALUE=[[1]]
IFD-PATH=[IFD] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[72/1]]
IFD-PATH=[IFD] ID=(0x011b) NAME=[YResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[72/1]]
IFD-PATH=[IFD] ID=(0x0128) NAME=[ResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[[2]]
IFD-PATH=[IFD] ID=(0x0132) NAME=[DateTime] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]
IFD-PATH=[IFD] ID=(0x013b) NAME=[Artist] COUNT=(1) TYPE=[ASCII] VALUE=[]
IFD-PATH=[IFD] ID=(0x0213) NAME=[YCbCrPositioning] COUNT=(1) TYPE=[SHORT] VALUE=[[2]]
IFD-PATH=[IFD] ID=(0x8298) NAME=[Copyright] COUNT=(1) TYPE=[ASCII] VALUE=[]
IFD-PATH=[IFD] ID=(0x8769) NAME=[ExifTag] COUNT=(1) TYPE=[LONG] VALUE=[[360]]
IFD-PATH=[IFD/Exif] ID=(0x829a) NAME=[ExposureTime] COUNT=(1) TYPE=[RATIONAL] VALUE=[[1/640]]
IFD-PATH=[IFD/Exif] ID=(0x829d) NAME=[FNumber] COUNT=(1) TYPE=[RATIONAL] VALUE=[[4/1]]
IFD-PATH=[IFD/Exif] ID=(0x8822) NAME=[ExposureProgram] COUNT=(1) TYPE=[SHORT] VALUE=[[4]]
IFD-PATH=[IFD/Exif] ID=(0x8827) NAME=[ISOSpeedRatings] COUNT=(1) TYPE=[SHORT] VALUE=[[1600]]
IFD-PATH=[IFD/Exif] ID=(0x8830) NAME=[SensitivityType] COUNT=(1) TYPE=[SHORT] VALUE=[[2]]
IFD-PATH=[IFD/Exif] ID=(0x8832) NAME=[RecommendedExposureIndex] COUNT=(1) TYPE=[LONG] VALUE=[[1600]]
IFD-PATH=[IFD/Exif] ID=(0x9000) NAME=[ExifVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[0230]
IFD-PATH=[IFD/Exif] ID=(0x9003) NAME=[DateTimeOriginal] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]
IFD-PATH=[IFD/Exif] ID=(0x9004) NAME=[DateTimeDigitized] COUNT=(20) TYPE=[ASCII] VALUE=[2017:12:02 08:18:50]
IFD-PATH=[IFD/Exif] ID=(0x9101) NAME=[ComponentsConfiguration] COUNT=(4) TYPE=[UNDEFINED] VALUE=[Exif9101ComponentsConfiguration<ID=[YCBCR] BYTES=[1 2 3 0]>]
IFD-PATH=[IFD/Exif] ID=(0x9201) NAME=[ShutterSpeedValue] COUNT=(1) TYPE=[SRATIONAL] VALUE=[[614400/65536]]
IFD-PATH=[IFD/Exif] ID=(0x9202) NAME=[ApertureValue] COUNT=(1) TYPE=[RATIONAL] VALUE=[[262144/65536]]
IFD-PATH=[IFD/Exif] ID=(0x9204) NAME=[ExposureBiasValue] COUNT=(1) TYPE=[SRATIONAL] VALUE=[[0/1]]
IFD-PATH=[IFD/Exif] ID=(0x9207) NAME=[MeteringMode] COUNT=(1) TYPE=[SHORT] VALUE=[[5]]
IFD-PATH=[IFD/Exif] ID=(0x9209) NAME=[Flash] COUNT=(1) TYPE=[SHORT] VALUE=[[16]]
IFD-PATH=[IFD/Exif] ID=(0x920a) NAME=[FocalLength] COUNT=(1) TYPE=[RATIONAL] VALUE=[[16/1]]
IFD-PATH=[IFD/Exif] ID=(0x927c) NAME=[MakerNote] COUNT=(8152) TYPE=[UNDEFINED] VALUE=[MakerNote<TYPE-ID=[28 00 01 00 03 00 31 00 00 00 74 05 00 00 02 00 03 00 04 00] LEN=(8152) SHA1=[d4154aa7df5474efe7ab38de2595919b9b4cc29f]>]
IFD-PATH=[IFD/Exif] ID=(0x9286) NAME=[UserComment] COUNT=(264) TYPE=[UNDEFINED] VALUE=[UserComment<SIZE=(256) ENCODING=[UNDEFINED] V=[0 0 0 0 0 0 0 0]... LEN=(256)>]
IFD-PATH=[IFD/Exif] ID=(0x9290) NAME=[SubSecTime] COUNT=(3) TYPE=[ASCII] VALUE=[00]
IFD-PATH=[IFD/Exif] ID=(0x9291) NAME=[SubSecTimeOriginal] COUNT=(3) TYPE=[ASCII] VALUE=[00]
IFD-PATH=[IFD/Exif] ID=(0x9292) NAME=[SubSecTimeDigitized] COUNT=(3) TYPE=[ASCII] VALUE=[00]
IFD-PATH=[IFD/Exif] ID=(0xa000) NAME=[FlashpixVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[0100]
IFD-PATH=[IFD/Exif] ID=(0xa001) NAME=[ColorSpace] COUNT=(1) TYPE=[SHORT] VALUE=[[1]]
IFD-PATH=[IFD/Exif] ID=(0xa002) NAME=[PixelXDimension] COUNT=(1) TYPE=[SHORT] VALUE=[[3840]]
IFD-PATH=[IFD/Exif] ID=(0xa003) NAME=[PixelYDimension] COUNT=(1) TYPE=[SHORT] VALUE=[[2560]]
IFD-PATH=[IFD/Exif] ID=(0xa005) NAME=[InteroperabilityTag] COUNT=(1) TYPE=[LONG] VALUE=[[9326]]
IFD-PATH=[IFD/Exif/Iop] ID=(0x0001) NAME=[InteroperabilityIndex] COUNT=(4) TYPE=[ASCII] VALUE=[R98]
IFD-PATH=[IFD/Exif/Iop] ID=(0x0002) NAME=[InteroperabilityVersion] COUNT=(4) TYPE=[UNDEFINED] VALUE=[0100]
IFD-PATH=[IFD/Exif] ID=(0xa20e) NAME=[FocalPlaneXResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[3840000/1461]]
IFD-PATH=[IFD/Exif] ID=(0xa20f) NAME=[FocalPlaneYResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[2560000/972]]
IFD-PATH=[IFD/Exif] ID=(0xa210) NAME=[FocalPlaneResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[[2]]
IFD-PATH=[IFD/Exif] ID=(0xa401) NAME=[CustomRendered] COUNT=(1) TYPE=[SHORT] VALUE=[[0]]
IFD-PATH=[IFD/Exif] ID=(0xa402) NAME=[ExposureMode] COUNT=(1) TYPE=[SHORT] VALUE=[[0]]
IFD-PATH=[IFD/Exif] ID=(0xa403) NAME=[WhiteBalance] COUNT=(1) TYPE=[SHORT] VALUE=[[0]]
IFD-PATH=[IFD/Exif] ID=(0xa406) NAME=[SceneCaptureType] COUNT=(1) TYPE=[SHORT] VALUE=[[0]]
IFD-PATH=[IFD/Exif] ID=(0xa430) NAME=[CameraOwnerName] COUNT=(1) TYPE=[ASCII] VALUE=[]
IFD-PATH=[IFD/Exif] ID=(0xa431) NAME=[BodySerialNumber] COUNT=(13) TYPE=[ASCII] VALUE=[063024020097]
IFD-PATH=[IFD/Exif] ID=(0xa432) NAME=[LensSpecification] COUNT=(4) TYPE=[RATIONAL] VALUE=[[16/1 35/1 0/1 0/1]]
IFD-PATH=[IFD/Exif] ID=(0xa434) NAME=[LensModel] COUNT=(22) TYPE=[ASCII] VALUE=[EF16-35mm f/4L IS USM]
IFD-PATH=[IFD/Exif] ID=(0xa435) NAME=[LensSerialNumber] COUNT=(11) TYPE=[ASCII] VALUE=[2400001068]
IFD-PATH=[IFD] ID=(0x8825) NAME=[GPSTag] COUNT=(1) TYPE=[LONG] VALUE=[[9554]]
IFD-PATH=[IFD/GPSInfo] ID=(0x0000) NAME=[GPSVersionID] COUNT=(4) TYPE=[BYTE] VALUE=[02 03 00 00]
IFD-PATH=[IFD1] ID=(0x0103) NAME=[Compression] COUNT=(1) TYPE=[SHORT] VALUE=[[6]]
IFD-PATH=[IFD1] ID=(0x011a) NAME=[XResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[72/1]]
IFD-PATH=[IFD1] ID=(0x011b) NAME=[YResolution] COUNT=(1) TYPE=[RATIONAL] VALUE=[[72/1]]
IFD-PATH=[IFD1] ID=(0x0128) NAME=[ResolutionUnit] COUNT=(1) TYPE=[SHORT] VALUE=[[2]]
IFD-PATH=[IFD1] ID=(0x0201) NAME=[JPEGInterchangeFormat] COUNT=(1) TYPE=[LONG] VALUE=[[11444]]
IFD-PATH=[IFD1] ID=(0x0202) NAME=[JPEGInterchangeFormatLength] COUNT=(1) TYPE=[LONG] VALUE=[[21491]]

`

	if actual != expected {
		t.Fatalf("Output not as expected:\nACTUAL:\n%s\nEXPECTED:\n%s", actual, expected)
	}
}

func TestMainJson(t *testing.T) {
	appFilepath := getAppFilepath()
	testImageFilepath := getTestImageFilepath()

	cmd := exec.Command(
		"go", "run", appFilepath,
		"--filepath", testImageFilepath,
		"--json")

	b := new(bytes.Buffer)
	cmd.Stdout = b
	cmd.Stderr = b

	err := cmd.Run()
	actualRaw := b.Bytes()

	if err != nil {
		fmt.Printf(string(actualRaw))
		log.Panic(err)
	}

	// Parse actual data.

	actual := make([]map[string]interface{}, 0)

	err = json.Unmarshal(actualRaw, &actual)
	log.PanicIf(err)

	// Read and parse expected data.

	assetsPath := exifcommon.GetTestAssetsPath()
	jsonFilepath := path.Join(assetsPath, "main_test_exif.json")

	expectedRaw, err := ioutil.ReadFile(jsonFilepath)
	log.PanicIf(err)

	expected := make([]map[string]interface{}, 0)

	err = json.Unmarshal(expectedRaw, &expected)
	log.PanicIf(err)

	// if reflect.DeepEqual(actual, expected) == false {
	// 	t.Fatalf("Output not as expected:\nACTUAL:\n%s\nEXPECTED:\n%s", actualRaw, expectedRaw)
	// }

	for i, tagInfo := range actual {
		if reflect.DeepEqual(tagInfo, expected[i]) == false {
			actualBytes, err := json.MarshalIndent(tagInfo, "", "  ")
			log.PanicIf(err)

			expectedBytes, err := json.MarshalIndent(expected[i], "", "  ")
			log.PanicIf(err)

			t.Fatalf("Tag (%d) not as expected:\nACTUAL:\n%s\nEXPECTED:\n%s", i, string(actualBytes), string(expectedBytes))
		}
	}

	if len(actual) != len(expected) {
		t.Fatalf("Actual tags not same length as expected tags.")
	}
}

func getAppFilepath() string {
	moduleRootPath := exifcommon.GetModuleRootPath()
	appFilepath := path.Join(moduleRootPath, "command", "exif-read-tool", "main.go")

	return appFilepath
}

func getTestImageFilepath() string {
	assetsPath := exifcommon.GetTestAssetsPath()
	testImageFilepath := path.Join(assetsPath, "NDM_8901.jpg")

	return testImageFilepath
}
