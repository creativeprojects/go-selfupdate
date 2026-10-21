// Get-release downloads the latest release of a repository and installs
// the binary into $GOPATH/bin, like "go get" would.
//
// Run it without arguments for usage.
package main

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/cmd"
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

	if help || flag.NArg() != 2 {
		usage()
		return
	}

	if verbose {
		selfupdate.SetLogger(log.New(os.Stdout, "", 0))
	}

	repo := flag.Arg(0)
	relExe := flag.Arg(1)

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

	ctx := context.Background()
	source, err := cmd.GetSource(ctx, cvsType, domain)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	latest, err := updater.DetectLatest(ctx, selfupdate.DetectLatestOpt{Repository: selfupdate.ParseSlug(slug)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cmd := getCommand(repo)
	cmdPath := filepath.Join(build.Default.GOPATH, "bin", cmd)

	if err := updater.UpdateTo(ctx, selfupdate.UpdateToOpt{Rel: latest, RelExe: relExe, CmdPath: cmdPath}); err != nil {
		fmt.Fprintf(os.Stderr, "Error while installing the release binary from %s: %s\n", latest.AssetURL, err)
		os.Exit(1)
	}

	fmt.Printf(`Command was updated to the latest version %s: %s

Release Notes:
%s
`, latest.Version(), cmdPath, latest.ReleaseNotes)
}

func usage() {
	fmt.Fprintln(os.Stderr, `
Usage: get-release [flags] {package} {relExe}

  get-release is like "go get github.com/owner/repo@latest".
  {package} is using the same format: "github.com/owner/repo".
  {relExe} is the name of the new executable file (usually inside a release archive).

Flags:`)
	flag.PrintDefaults()
}

func getCommand(pkg string) string {
	pkg = strings.TrimSuffix(pkg, "/")
	_, cmd := filepath.Split(pkg)
	return cmd
}
