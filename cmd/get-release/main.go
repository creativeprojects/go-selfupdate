package main

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/cmd"
)

func main() {
	var help, verbose bool
	var cvsType, baseURL string
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&verbose, "v", false, "Display debugging information")
	flag.StringVar(&cvsType, "t", "auto", "Version control: \"github\", \"gitea\", \"gitlab\" or \"http\"")
	flag.StringVar(&baseURL, "u", "", "Base URL for VCS on http or dedicated instances")

	flag.Usage = usage
	flag.Parse()

	if help || flag.NArg() != 1 {
		usage()
		return
	}

	if verbose {
		selfupdate.SetLogger(log.New(os.Stdout, "", 0))
	}

	repo := flag.Arg(0)

	domain, slug, err := cmd.SplitDomainSlug(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if domain == "" && baseURL != "" {
		domain = baseURL
	}

	if verbose {
		fmt.Printf("slug %q on domain %q\n", slug, domain)
	}

	source, err := cmd.GetSource(cvsType, domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	latest, found, err := updater.DetectLatest(ctx, selfupdate.ParseSlug(slug))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while detecting the latest version:", err)
		os.Exit(1)
	}
	if !found {
		fmt.Fprintln(os.Stderr, "No release found in", slug)
		os.Exit(1)
	}

	cmd := getCommand(flag.Arg(0))
	cmdPath := filepath.Join(build.Default.GOPATH, "bin", cmd)
	if _, err := os.Stat(cmdPath); err != nil {
		// When executable is not existing yet
		if err := installFrom(ctx, latest.AssetURL, cmd, cmdPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error while installing the release binary from %s: %s\n", latest.AssetURL, err)
			os.Exit(1)
		}
	} else {
		if err := updater.UpdateTo(ctx, latest, cmdPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error while replacing the binary with %s: %s\n", latest.AssetURL, err)
			os.Exit(1)
		}
	}

	fmt.Printf(`Command was updated to the latest version %s: %s

Release Notes:
%s
`, latest.Version(), cmdPath, latest.ReleaseNotes)
}

func usage() {
	fmt.Fprintln(os.Stderr, `
Usage: get-release [flags] {package}

  get-release is like "go get github.com/owner/repo@latest".
  {package} is using the same format: "github.com/owner/repo".

Flags:`)
	flag.PrintDefaults()
}

func getCommand(pkg string) string {
	pkg = strings.TrimSuffix(pkg, "/")
	_, cmd := filepath.Split(pkg)
	return cmd
}

func installFrom(ctx context.Context, url, cmd, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request to download release binary from %s: %s", url, err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download release binary from %s: %s", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download release binary from %s: Invalid response ", url)
	}
	executable, err := selfupdate.DecompressCommand(res.Body, url, cmd, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("failed to decompress downloaded asset from %s: %s", url, err)
	}
	bin, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(bin, executable); err != nil {
		return fmt.Errorf("failed to write binary to %s: %s", path, err)
	}
	return nil
}
