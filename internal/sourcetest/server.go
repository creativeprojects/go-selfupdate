package sourcetest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
)

// startServer serves handler until the end of the test.
func startServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// repositoryExists reports whether data serves the owner/repo repository.
func repositoryExists(data *memory.Source, owner, repo string) bool {
	if data.NotFound {
		return false
	}
	if data.Repository == nil {
		return true
	}
	expectedOwner, expectedRepo, err := data.Repository.GetSlug()
	return err == nil && owner == expectedOwner && repo == expectedRepo
}

// findRelease returns the release with this ID.
func findRelease(data *memory.Source, id int64) (selfupdate.SourceRelease, bool) {
	for _, release := range data.Releases {
		if release.GetID() == id {
			return release, true
		}
	}
	return nil, false
}

// findAsset returns the asset with this ID from any release.
func findAsset(data *memory.Source, id int64) (selfupdate.SourceAsset, bool) {
	for _, release := range data.Releases {
		if asset, ok := findReleaseAsset(release, id); ok {
			return asset, true
		}
	}
	return nil, false
}

// findReleaseAsset returns the asset with this ID from release.
func findReleaseAsset(release selfupdate.SourceRelease, id int64) (selfupdate.SourceAsset, bool) {
	for _, asset := range release.GetAssets() {
		if asset.GetID() == id {
			return asset, true
		}
	}
	return nil, false
}

// pathID returns the numeric path wildcard called name, or -1 when it is not a number.
func pathID(r *http.Request, name string) int64 {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		return -1
	}
	return id
}

// downloadURL is where the fake server receiving r serves the content of an asset.
func downloadURL(r *http.Request, id int64) string {
	return fmt.Sprintf("http://%s/download/%d", r.Host, id)
}

// handleDownload registers the route serving asset content.
func handleDownload(mux *http.ServeMux, data *memory.Source) {
	mux.HandleFunc("GET /download/{id}", func(w http.ResponseWriter, r *http.Request) {
		writeFile(w, data, pathID(r, "id"))
	})
}

func writeFile(w http.ResponseWriter, data *memory.Source, id int64) {
	content, ok := data.Files[id]
	if !ok {
		writeNotFound(w)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(content)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

// writeNotFound sends the error body the GitHub, GitLab and Gitea SDKs all decode.
func writeNotFound(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"message":"Not Found"}`))
}
