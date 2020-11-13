package selfupdate

type SourceRelease interface {
	GetTagName() string
	GetDraft() bool
	GetPrerelease() bool
}

// Source interface to load the releases from (GitHubSource for example)
type Source interface {
	ListReleases(owner, repo string) ([]SourceRelease, error)
}
