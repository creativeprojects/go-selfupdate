// Package selfupdate detects the latest release of a program on a source
// provider (GitHub, GitLab, Gitea or a plain HTTP server) and, when it is
// newer than the running version, downloads the released binary and replaces
// the running executable with it.
//
//   - Automatically detects the latest version of the released binary
//   - Retrieves the proper binary for the OS and architecture it runs on
//   - Updates the binary with rollback support on failure
//   - Tested on Linux, macOS and Windows
//   - Supports the different versions of the ARM architecture
//   - Supports macOS universal binaries
//   - Supports many archive and compression formats (zip, tar, gzip, xz, bzip2)
//   - Supports private repositories
//   - Supports hash and signature validation
//
// There is no default source: [Config.Source] must be set to one from the
// provider packages (github, gitlab, gitea or http). Create an [Updater]
// with [NewUpdater], then call [Updater.DetectLatest] and [Updater.UpdateTo],
// or [Updater.UpdateSelf] to do both at once.
//
// Released binaries and git tags must follow the naming rules described in the README:
//
//   - https://github.com/creativeprojects/go-selfupdate#naming-rules-of-released-binaries
//   - https://github.com/creativeprojects/go-selfupdate#naming-rules-of-versions-git-tags
//
// Two command line tools wrapping this library are available:
//
//   - https://github.com/creativeprojects/go-selfupdate/tree/main/cmd/detect-latest-release
//   - https://github.com/creativeprojects/go-selfupdate/tree/main/cmd/get-release
package selfupdate
