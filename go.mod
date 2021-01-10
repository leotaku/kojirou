module github.com/leotaku/manki

go 1.15

require (
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/fatih/color v1.10.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/leotaku/mobi v0.0.0-20210110091135-443c0ffddbe9
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/sys v0.0.0-20210110051926-789bb1bd4061 // indirect
	golang.org/x/text v0.3.5
)

replace github.com/leotaku/mobi => ../mobi
