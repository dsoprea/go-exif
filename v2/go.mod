module github.com/dsoprea/go-exif/v2

go 1.13

// Development only
replace github.com/dsoprea/go-logging => ../../go-logging

require (
	github.com/dsoprea/go-logging v0.0.0-20200517223158-a10564966e9d
	github.com/golang/geo v0.0.0-20200319012246-673a6f80352d
	github.com/jessevdk/go-flags v1.4.0 // indirect
	golang.org/x/net v0.0.0-20200513185701-a91f0712d120 // indirect
	gopkg.in/yaml.v2 v2.3.0
)
