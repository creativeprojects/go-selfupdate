package sourcetest

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/creativeprojects/go-selfupdate/v2/memory"
	yaml "go.yaml.in/yaml/v3"
)

// manifest mirrors the YAML document read by the http source.
type manifest struct {
	LastReleaseID int64              `yaml:"last_release_id"`
	LastAssetID   int64              `yaml:"last_asset_id"`
	Releases      []*manifestRelease `yaml:"releases"`
}

type manifestRelease struct {
	ID           int64            `yaml:"id"`
	Name         string           `yaml:"name"`
	TagName      string           `yaml:"tag_name"`
	URL          string           `yaml:"url"`
	Draft        bool             `yaml:"draft"`
	Prerelease   bool             `yaml:"prerelease"`
	PublishedAt  time.Time        `yaml:"published_at"`
	ReleaseNotes string           `yaml:"release_notes"`
	Assets       []*manifestAsset `yaml:"assets"`
}

type manifestAsset struct {
	ID   int64  `yaml:"id"`
	Name string `yaml:"name"`
	Size int    `yaml:"size"`
	URL  string `yaml:"url"`
}

// NewHTTPServer starts a fake HTTP repository serving data.
// Use its URL as http.Config.BaseURL.
func NewHTTPServer(t *testing.T, data *memory.Source) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{owner}/{repo}/manifest.yaml", func(w http.ResponseWriter, r *http.Request) {
		if !repositoryExists(data, r.PathValue("owner"), r.PathValue("repo")) {
			writeNotFound(w)
			return
		}
		content, err := yaml.Marshal(newManifest(r, data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(content)
	})
	handleDownload(mux, data)
	return startServer(t, mux)
}

func newManifest(r *http.Request, data *memory.Source) *manifest {
	result := &manifest{
		Releases: make([]*manifestRelease, 0, len(data.Releases)),
	}
	if data.LatestRelease != nil {
		result.LastReleaseID = data.LatestRelease.GetID()
	}
	for _, release := range data.Releases {
		assets := make([]*manifestAsset, 0, len(release.GetAssets()))
		for _, asset := range release.GetAssets() {
			assets = append(assets, &manifestAsset{
				ID:   asset.GetID(),
				Name: asset.GetName(),
				Size: asset.GetSize(),
				URL:  downloadURL(r, asset.GetID()),
			})
			result.LastAssetID = max(result.LastAssetID, asset.GetID())
		}
		result.Releases = append(result.Releases, &manifestRelease{
			ID:           release.GetID(),
			Name:         release.GetName(),
			TagName:      release.GetTagName(),
			URL:          release.GetURL(),
			Draft:        release.GetDraft(),
			Prerelease:   release.GetPrerelease(),
			PublishedAt:  release.GetPublishedAt(),
			ReleaseNotes: release.GetReleaseNotes(),
			Assets:       assets,
		})
	}
	return result
}
