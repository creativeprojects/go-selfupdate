package selfupdate

type GitHubRelease struct {
	TagName string
}

func (r *GitHubRelease) GetTagName() string {
	return r.TagName
}

// Verify interface
var _ SourceRelease = &GitHubRelease{}
