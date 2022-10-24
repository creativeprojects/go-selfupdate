package selfupdate

import (
	"fmt"
	"regexp"
	"runtime"
)

// Updater is responsible for managing the context of self-update.
type Updater struct {
	source     Source
	validator  Validator
	filters    []*regexp.Regexp
	os         string
	arch       string
	arm        uint8
	prerelease bool
	draft      bool
}

// keep the default updater instance in cache
var defaultUpdater *Updater

// NewUpdater creates a new updater instance.
// If you don't specify a source in the config object, GitHub will be used
func NewUpdater(config Config) (*Updater, error) {
	source := config.Source
	if source == nil {
		// default source is GitHub
		source, _ = NewGitHubSource(GitHubConfig{})
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
	arch := config.Arch
	if os == "" {
		os = runtime.GOOS
	}
	if arch == "" {
		arch = runtime.GOARCH
	}
	arm := config.Arm
	if arm == 0 && goarm > 0 {
		arm = goarm
	}

	return &Updater{
		source:     source,
		validator:  config.Validator,
		filters:    filtersRe,
		os:         os,
		arch:       arch,
		arm:        arm,
		prerelease: config.Prerelease,
		draft:      config.Draft,
	}, nil
}

// DefaultUpdater creates a new updater instance with default configuration.
// It initializes GitHub API client with default API base URL.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
// Every call to this function will always return the same instance, it's only created once
func DefaultUpdater() *Updater {
	// instantiate it only once
	if defaultUpdater != nil {
		return defaultUpdater
	}
	// an error can only be returned when using GitHub Enterprise URLs
	// so we're safe here :)
	source, _ := NewGitHubSource(GitHubConfig{})
	defaultUpdater = &Updater{
		source: source,
		os:     runtime.GOOS,
		arch:   runtime.GOARCH,
		arm:    goarm,
	}
	return defaultUpdater
}
