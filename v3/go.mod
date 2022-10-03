module github.com/dsoprea/go-exif/v3

go 1.12

// Development only
// replace github.com/dsoprea/go-logging => ../../go-logging
// replace github.com/dsoprea/go-utility/v2 => ../../go-utility/v2

require (
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd
	github.com/dsoprea/go-utility/v2 v2.0.0-20221003160719-7bc88537c05e
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/jessevdk/go-flags v1.5.0
	golang.org/x/net v0.0.0-20221002022538-bcab6841153b // indirect
	golang.org/x/sys v0.0.0-20220928140112-f11e5e49a4ec // indirect
	gopkg.in/yaml.v2 v2.4.0
)
