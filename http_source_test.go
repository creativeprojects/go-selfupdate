// Copyright (c) 2024 Mr. Gecko's Media (James Coleman). http://mrgeckosmedia.com/
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package selfupdate

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	goLog "log"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unused port for testing.
const httpRepoTestPort = 8831

// Test server for testing http repos.
type HttpRepoTestServer struct {
	server  *http.Server
	repoURL string
}

// Setup test server with test data.
func NewHttpRepoTestServer() *HttpRepoTestServer {
	s := new(HttpRepoTestServer)

	// Setup handlers.
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./testdata/http_repo"))
	mux.Handle("/repo/creativeprojects/resticprofile/", http.StripPrefix("/repo/creativeprojects/resticprofile", fs))

	// Setup server config.
	srvAddr := fmt.Sprintf("127.0.0.1:%d", httpRepoTestPort)
	s.server = &http.Server{
		Addr:    srvAddr,
		Handler: mux,
	}
	s.repoURL = "http://" + srvAddr + "/repo/"
	return s
}

// Run the http server.
func (s *HttpRepoTestServer) Run() {
	isListening := make(chan bool)
	// Start server.
	go s.Start(isListening)
	// Allow the http server to initialize.
	<-isListening
}

// Stop the HTTP server.
func (s *HttpRepoTestServer) Stop() {
	s.server.Shutdown(context.Background())
}

// Start the HTTP server with a notification channel
// for when the server is listening.
func (s *HttpRepoTestServer) Start(isListening chan bool) {
	// Start server.
	l, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		goLog.Fatal("Listen: ", err)
	}
	// Now notify we are listening.
	isListening <- true
	// Serve http server on the listening port.
	err = s.server.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		goLog.Fatal("Serve: ", err)
	}
}

// Verify the client ignores invalid URLs.
func TestHttpClientInvalidURL(t *testing.T) {
	_, err := NewHttpSource(HttpConfig{BaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

// Verify the client accepts valid URLs.
func TestHttpClientValidURL(t *testing.T) {
	_, err := NewHttpSource(HttpConfig{BaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitHub source with valid URL")
	}
}

// Verify cancelled contexts actually cancels a request.
func TestHttpListReleasesContextCancelled(t *testing.T) {
	// Make a valid HTTP source.
	source, err := NewHttpSource(HttpConfig{BaseURL: "http://localhost"})
	require.NoError(t, err)

	// Create a cancelled context.
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	// Attempt to list releases and verify result.
	_, err = source.ListReleases(ctx, ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

// Verify cancelled contexts actually cancels a download.
func TestHttpDownloadReleaseAssetContextCancelled(t *testing.T) {
	// Make a valid HTTP source.
	source, err := NewHttpSource(HttpConfig{BaseURL: "http://localhost"})
	require.NoError(t, err)

	// Create a cancelled context.
	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	// Attempt to download release and verify result.
	_, err = source.DownloadReleaseAsset(ctx, &Release{
		AssetID:  11,
		AssetURL: "http://localhost/",
	}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

// Verify no release actually returns an error.
func TestHttpDownloadReleaseAssetWithNilRelease(t *testing.T) {
	// Create valid HTTP source.
	source, err := NewHttpSource(HttpConfig{BaseURL: "http://localhost"})
	require.NoError(t, err)

	// Attempt to download release without specifying the release and verify result.
	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, ErrInvalidRelease)
}

// Verify we're able to list releases and download an asset.
func TestHttpListAndDownloadReleaseAsset(t *testing.T) {
	// Create test HTTP server and start it.
	server := NewHttpRepoTestServer()
	server.Run()

	// Make HTTP source with our test server.
	source, err := NewHttpSource(HttpConfig{BaseURL: server.repoURL})
	require.NoError(t, err)

	// List releases
	releases, err := source.ListReleases(context.Background(), ParseSlug("creativeprojects/resticprofile"))
	require.NoError(t, err)

	// Confirm the manifest parsed the correct number of releases.
	if len(releases) != 2 {
		t.Fatal("releases count is not valid")
	}

	// Confirm the manifest parsed by the first release is valid.
	if releases[0].GetTagName() != "v0.1.1" {
		t.Fatal("release is not as expected")
	}

	// Confirm the release assets are parsed correctly.
	assets := releases[1].GetAssets()
	if assets[1].GetName() != "example_linux_amd64.tar.gz" {
		t.Fatal("the release asset is not valid")
	}

	// Get updater with source.
	updater, err := NewUpdater(Config{
		Source: source,
		OS:     "linux",
		Arch:   "amd64",
	})
	require.NoError(t, err)

	// Find the latest release.
	release, found, err := updater.DetectLatest(context.Background(), NewRepositorySlug("creativeprojects", "resticprofile"))
	require.NoError(t, err)
	if !found {
		t.Fatal("no release found")
	}

	// Download asset.
	body, err := source.DownloadReleaseAsset(context.Background(), release, 5)
	require.NoError(t, err)

	// Read data.
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// Verify data.
	hfun := md5.New()
	hfun.Write(data)
	sum := hfun.Sum(nil)
	hash := hex.EncodeToString(sum)
	if hash != "9cffcbe826ae684db1c8a08ff9216f34" {
		t.Errorf("hash isn't valid for test file: %s", hash)
	}

	// Stop as we're done.
	server.Stop()
}
