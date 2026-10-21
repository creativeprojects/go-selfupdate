package sourcetest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// NewGitLabServer starts a fake GitLab API serving data.
// Use its URL as gitlab.Config.BaseURL.
func NewGitLabServer(t *testing.T, data *memory.Source) *httptest.Server {
	t.Helper()

	// the project is a URL-encoded "owner/repo" slug, decoded by the router
	projectExists := func(r *http.Request) bool {
		owner, repo, err := selfupdate.ParseSlug(r.PathValue("project")).GetSlug()
		return err == nil && repositoryExists(data, owner, repo)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v4/projects/{project}", func(w http.ResponseWriter, r *http.Request) {
		if !projectExists(r) {
			writeNotFound(w)
			return
		}
		writeJSON(w, &gitlab.Project{PathWithNamespace: r.PathValue("project")})
	})
	mux.HandleFunc("GET /api/v4/projects/{project}/releases", func(w http.ResponseWriter, r *http.Request) {
		if !projectExists(r) {
			writeNotFound(w)
			return
		}
		releases := make([]*gitlab.Release, 0, len(data.Releases))
		for _, release := range data.Releases {
			releases = append(releases, gitlabRelease(r, release))
		}
		writeJSON(w, releases)
	})
	mux.HandleFunc("GET /api/v4/projects/{project}/releases/permalink/latest", func(w http.ResponseWriter, r *http.Request) {
		if !projectExists(r) || data.LatestRelease == nil {
			writeNotFound(w)
			return
		}
		writeJSON(w, gitlabRelease(r, data.LatestRelease))
	})
	handleDownload(mux, data)
	return startServer(t, mux)
}

// gitlabRelease converts a release: GitLab has no release ID, draft or pre-release flag, and no asset size.
func gitlabRelease(r *http.Request, release selfupdate.SourceRelease) *gitlab.Release {
	links := make([]*gitlab.ReleaseLink, 0, len(release.GetAssets()))
	for _, asset := range release.GetAssets() {
		links = append(links, &gitlab.ReleaseLink{
			ID:             asset.GetID(),
			Name:           asset.GetName(),
			URL:            downloadURL(r, asset.GetID()),
			DirectAssetURL: downloadURL(r, asset.GetID()),
		})
	}
	releasedAt := release.GetPublishedAt()
	return &gitlab.Release{
		TagName:     release.GetTagName(),
		Name:        release.GetName(),
		Description: release.GetReleaseNotes(),
		ReleasedAt:  &releasedAt,
		Commit:      gitlab.Commit{WebURL: release.GetURL()},
		Assets: gitlab.ReleaseAssets{
			Count: int64(len(links)),
			Links: links,
		},
	}
}
