Download and install the latest release of a command line tool from GitHub, Gitea, GitLab or an HTTP repository.

## Installation

```
$ go install github.com/creativeprojects/go-selfupdate/v2/cmd/get-release@latest
```

## Usage

Usage is quite similar to `go get`.

```
$ get-release [flags] {package} {relExe}
```

`{package}` is the repository, in the same format as `go get`: `github.com/owner/repo`.
`{relExe}` is the name of the executable inside the release archive, usually the name of the command.

Flags:

- `-t`: version control, one of `github`, `gitea`, `gitlab` or `http`. The default `auto` picks `gitea` or `gitlab` when the domain name contains it, and `github` otherwise.
- `-u`: base URL of the server when `{package}` is a plain `owner/repo` slug, for dedicated instances and HTTP repositories.
- `-v`: display debugging information.

Please note that this command assumes that specified package is following Git tag naming rules and
released binaries naming rules described in [README](../../README.md).

For example, following command downloads and installs the released binary of [ghr](https://github.com/tcnksm/ghr)
to `$GOPATH/bin`.

```
$ get-release github.com/tcnksm/ghr ghr
Command was updated to the latest version 0.5.4: /Users/you/.go/bin/ghr

$ ghr -version
ghr version v0.5.4 (a12ff1c)
```
