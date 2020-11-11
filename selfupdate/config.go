package selfupdate

import "context"

// Config represents the configuration of self-update.
type Config struct {
	// APIToken represents GitHub API token. If it's not empty, it will be used for authentication of GitHub API
	APIToken string
	// EnterpriseBaseURL is a base URL of GitHub API. If you want to use this library with GitHub Enterprise,
	// please set "https://{your-organization-address}/api/v3/" to this field.
	EnterpriseBaseURL string
	// EnterpriseUploadURL is a URL to upload stuffs to GitHub Enterprise instance. This is often the same as an API base URL.
	// So if this field is not set and EnterpriseBaseURL is set, EnterpriseBaseURL is also set to this field.
	EnterpriseUploadURL string
	// Validator represents types which enable additional validation of downloaded release.
	Validator Validator
	// Filters are regexp used to filter on specific assets for releases with multiple assets.
	// An asset is selected if it matches any of those, in addition to the regular tag, os, arch, extensions.
	// Please make sure that your filter(s) uniquely match an asset.
	Filters []string
	// Context used by the http client (default to context.Background)
	Context context.Context
	// OS is set to the value of runtime.GOOS by default, but you can force another value here
	OS string
	// Arch is set to the value of runtime.GOARCH by default, but you can force another value here
	Arch string
	// Arm 32bits version. Valid values are 0 (unknown), 5, 6 or 7. Default is detected value (if any)
	Arm uint8
}
