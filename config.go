package selfupdate

// Config represents the configuration of self-update.
type Config struct {
	// Source where to load the releases from (example: GitHubSource)
	Source Source
	// Validator represents types which enable additional validation of downloaded release.
	Validator Validator
	// Filters are regexp used to filter on specific assets for releases with multiple assets.
	// An asset is selected if it matches any of those, in addition to the regular tag, os, arch, extensions.
	// Please make sure that your filter(s) uniquely match an asset.
	Filters []string
	// OS is set to the value of runtime.GOOS by default, but you can force another value here
	OS string
	// Arch is set to the value of runtime.GOARCH by default, but you can force another value here
	Arch string
	// Arm 32bits version. Valid values are 0 (unknown), 5, 6 or 7. Default is detected value (if any)
	Arm uint8
	// Draft permits an upgrade to a "draft" version (default to false)
	Draft bool
	// Prerelease permits an upgrade to a "pre-release" version (default to false)
	Prerelease bool
}
