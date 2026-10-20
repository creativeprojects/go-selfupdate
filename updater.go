package selfupdate

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/creativeprojects/go-selfupdate/internal"
	"github.com/creativeprojects/go-selfupdate/types"
)

// Updater is responsible for managing the context of self-update.
type Updater struct {
	source        types.Source
	validator     types.Validator
	filters       []*regexp.Regexp
	os            string
	arch          string
	arm           uint8
	universalArch string // only filled in when needed
	prerelease    bool
	draft         bool
	oldSavePath   string
}

// keep the default updater instance in cache
var defaultUpdater *Updater

// NewUpdater creates a new updater instance.
// Returns an error if the source is not specified in the config.
func NewUpdater(config types.Config) (*Updater, error) {
	source := config.Source
	if source == nil {
		return nil, types.ErrInvalidSource
	}

	filtersRe := make([]*regexp.Regexp, 0, len(config.Filters))
	for _, filter := range config.Filters {
		re, err := regexp.Compile(filter)
		if err != nil {
			return nil, fmt.Errorf("could not compile regular expression %q for filtering releases: %w", filter, err)
		}
		filtersRe = append(filtersRe, re)
	}

	os := config.OS
	if os == "" {
		os = runtime.GOOS
	}
	arch := config.Arch
	if arch == "" {
		arch = runtime.GOARCH
	}
	arm := config.Arm
	if arm == 0 && arch == "arm" {
		exe, _ := internal.GetExecutablePath()
		arm = getGOARM(exe)
	}
	universalArch := ""
	if os == "darwin" && config.UniversalArch != "" {
		universalArch = config.UniversalArch
	}

	return &Updater{
		source:        source,
		validator:     config.Validator,
		filters:       filtersRe,
		os:            os,
		arch:          arch,
		arm:           arm,
		universalArch: universalArch,
		prerelease:    config.Prerelease,
		draft:         config.Draft,
		oldSavePath:   config.OldSavePath,
	}, nil
}
