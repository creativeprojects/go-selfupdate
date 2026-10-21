package sourcetest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	"github.com/google/go-github/v93/github"
)

// NewGitHubServer starts a fake GitHub API serving data.
// Use its URL as github.Config.EnterpriseBaseURL.
func NewGitHubServer(t *testing.T, data *memory.Source) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v3/repos/{owner}/{repo}/releases", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) {
			writeNotFound(w)
			return
		}
		releases := make([]*github.RepositoryRelease, 0, len(data.Releases))
		for _, release := range data.Releases {
			releases = append(releases, githubRelease(r, release))
		}
		writeJSON(w, releases)
	})
	mux.HandleFunc("GET /api/v3/repos/{owner}/{repo}/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) || data.LatestRelease == nil {
			writeNotFound(w)
			return
		}
		writeJSON(w, githubRelease(r, data.LatestRelease))
	})
	mux.HandleFunc("GET /api/v3/repos/{owner}/{repo}/releases/assets/{id}", func(w http.ResponseWriter, r *http.Request) {
		asset, ok := findAsset(data, pathID(r, "id"))
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) || !ok {
			writeNotFound(w)
			return
		}
		// the API sends the content instead of the metadata when asked for a binary
		if strings.Contains(r.Header.Get("Accept"), "application/octet-stream") {
			writeFile(w, data, asset.GetID())
			return
		}
		writeJSON(w, githubAsset(r, asset))
	})
	handleDownload(mux, data)
	return startServer(t, mux)
}

func githubRelease(r *http.Request, release selfupdate.SourceRelease) *github.RepositoryRelease {
	assets := make([]*github.ReleaseAsset, 0, len(release.GetAssets()))
	for _, asset := range release.GetAssets() {
		assets = append(assets, githubAsset(r, asset))
	}
	return &github.RepositoryRelease{
		ID:          release.GetID(),
		TagName:     release.GetTagName(),
		Name:        new(release.GetName()),
		Body:        new(release.GetReleaseNotes()),
		HTMLURL:     release.GetURL(),
		Draft:       release.GetDraft(),
		Prerelease:  release.GetPrerelease(),
		PublishedAt: &github.Timestamp{Time: release.GetPublishedAt()},
		Assets:      assets,
	}
}

func githubAsset(r *http.Request, asset selfupdate.SourceAsset) *github.ReleaseAsset {
	return &github.ReleaseAsset{
		ID:                 new(asset.GetID()),
		Name:               new(asset.GetName()),
		Size:               new(asset.GetSize()),
		BrowserDownloadURL: new(downloadURL(r, asset.GetID())),
	}
}
