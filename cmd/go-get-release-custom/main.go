package main

import (
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
)

func main() {
	var help, verbose bool
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&verbose, "v", false, "Display debugging information")

	flag.Usage = usage
	flag.Parse()

	if help || flag.NArg() != 1 {
		usage()
		os.Exit(1)
	}

	if verbose {
		selfupdate.SetLogger(log.New(os.Stdout, "", 0))
	}

	source, _ := selfupdate.NewGiteaSource(selfupdate.GiteaConfig{BaseURL: "https://gitea.com/"})
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: nil,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Arm:       0,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	latest, found, err := updater.DetectLatest(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while detecting the latest version:", err)
		os.Exit(1)
	}
	if !found {
		fmt.Fprintln(os.Stderr, "No release was found in", flag.Arg(0))
		os.Exit(1)
	}

	cmd := getCommand(flag.Arg(0))
	cmdPath := filepath.Join(build.Default.GOPATH, "bin", cmd)
	if _, err := os.Stat(cmdPath); err != nil {
		// When executable is not existing yet
		if err := installFrom(latest.AssetURL, cmd, cmdPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error while installing the release binary from %s: %s\n", latest.AssetURL, err)
			os.Exit(1)
		}
	} else {
		if err := updater.UpdateTo(latest, cmdPath); err != nil {
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
	fmt.Fprintln(os.Stderr, `Usage: go-get-release [flags] {package}

  go-get-release is like "go get", but it downloads the latest release from your git provider.
  Specify the slug (<owner>/<repo>) as the fiorst argument to this command

Flags:`)
	flag.PrintDefaults()
}

func getCommand(pkg string) string {
	_, cmd := filepath.Split(pkg)
	if cmd == "" {
		// When pkg path is ending with path separator, we need to split it out.
		// i.e. github.com/creativeprojects/foo/cmd/bar/
		_, cmd = filepath.Split(cmd)
	}
	return cmd
}

func installFrom(url, cmd, name, path string) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download release binary from %s: %s", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("failed to download release binary from %s: Invalid response ", url)
	}
	executable, err := selfupdate.DecompressCommand(res.Body, url, name, cmd, runtime.GOOS, runtime.GOARCH)
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
