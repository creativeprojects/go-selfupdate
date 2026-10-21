package sourcetest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/memory"
)

// NewGiteaServer starts a fake Gitea API serving data.
// Use its URL as gitea.Config.BaseURL.
func NewGiteaServer(t *testing.T, data *memory.Source) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	// the SDK checks the server version when creating a client
	mux.HandleFunc("GET /api/v1/version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"version": "1.24.0"})
	})
	mux.HandleFunc("GET /api/v1/repos/{owner}/{repo}/releases", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) {
			writeNotFound(w)
			return
		}
		releases := make([]*gitea.Release, 0, len(data.Releases))
		for _, release := range data.Releases {
			releases = append(releases, giteaRelease(r, release))
		}
		writeJSON(w, releases)
	})
	mux.HandleFunc("GET /api/v1/repos/{owner}/{repo}/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) || data.LatestRelease == nil {
			writeNotFound(w)
			return
		}
		writeJSON(w, giteaRelease(r, data.LatestRelease))
	})
	mux.HandleFunc("GET /api/v1/repos/{owner}/{repo}/releases/{release}/assets/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) {
			writeNotFound(w)
			return
		}
		release, ok := findRelease(data, pathID(r, "release"))
		if !ok {
			writeNotFound(w)
			return
		}
		asset, ok := findReleaseAsset(release, pathID(r, "id"))
		if !ok {
			writeNotFound(w)
			return
		}
		writeJSON(w, giteaAttachment(r, asset))
	})
	handleDownload(mux, data)
	return startServer(t, mux)
}

func giteaRelease(r *http.Request, release selfupdate.SourceRelease) *gitea.Release {
	attachments := make([]*gitea.Attachment, 0, len(release.GetAssets()))
	for _, asset := range release.GetAssets() {
		attachments = append(attachments, giteaAttachment(r, asset))
	}
	return &gitea.Release{
		ID:           release.GetID(),
		TagName:      release.GetTagName(),
		Title:        release.GetName(),
		Note:         release.GetReleaseNotes(),
		HTMLURL:      release.GetURL(),
		IsDraft:      release.GetDraft(),
		IsPrerelease: release.GetPrerelease(),
		PublishedAt:  release.GetPublishedAt(),
		Attachments:  attachments,
	}
}

func giteaAttachment(r *http.Request, asset selfupdate.SourceAsset) *gitea.Attachment {
	return &gitea.Attachment{
		ID:          asset.GetID(),
		Name:        asset.GetName(),
		Size:        int64(asset.GetSize()),
		DownloadURL: downloadURL(r, asset.GetID()),
	}
}
