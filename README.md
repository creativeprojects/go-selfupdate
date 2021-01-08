Self-Update library for GitHub hosted applications in Go
========================================================

[![Godoc reference](https://godoc.org/github.com/creativeprojects/go-selfupdate?status.svg)](http://godoc.org/github.com/creativeprojects/go-selfupdate)
[![Build](https://github.com/creativeprojects/go-selfupdate/workflows/Build/badge.svg)](https://github.com/creativeprojects/go-selfupdate/actions)
[![codecov](https://codecov.io/gh/creativeprojects/go-selfupdate/branch/main/graph/badge.svg?token=3FejM0fkw2)](https://codecov.io/gh/creativeprojects/go-selfupdate)

go-selfupdate detects the information of the latest release via [GitHub Releases API][] and
checks the current version. If a newer version than itself is detected, it downloads the released binary from
GitHub and replaces itself.

- Automatically detect the latest version of released binary on GitHub
- Retrieve the proper binary for the OS and arch where the binary is running
- Update the binary with rollback support on failure
- Tested on Linux, macOS and Windows
- Many archive and compression formats are supported (zip, tar, gzip, xzip, bzip2)
- Support private repositories
- Support hash, signature validation

[GitHub Releases API]: https://developer.github.com/v3/repos/releases/

This library started as a fork of https://github.com/rhysd/go-github-selfupdate. A few things have changed from the original implementation:
- don't expose an external semver.Version type, but provide the same functionality through the API: LessThan, Equal and GreaterThan
- use an interface to send logs (compatible with standard log.Logger)
- able to detect different ARM CPU architectures (the original library wasn't working on my different versions of raspberry pi)
- support for assets compressed with bzip2 (.bz2)
- can use a single file containing the sha256 checksums for all the files (one per line)

### Example

Here's an example how to use the library for an application to update itself

```go
func update(version string) error {
	latest, found, err := selfupdate.DetectLatest("creativeprojects/resticprofile")
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %v", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		log.Printf("Current version (%s) is the latest", version)
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %v", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
	return nil
}
```

### Important note

The API can change anytime until it reaches version 1.0.
It is unlikely it will change drastically though, but it can.

### Naming Rules of Released Binaries

go-selfupdate assumes that released binaries are put for each combination of platforms and architectures.
Binaries for each platform can be easily built using tools like [goreleaser][]

You need to put the binaries with the following format.

```
{cmd}_{goos}_{goarch}{.ext}
```

`{cmd}` is a name of command.
`{goos}` and `{goarch}` are the platform and the arch type of the binary.
`{.ext}` is a file extension. go-selfupdate supports `.zip`, `.gzip`, `.bz2`, `.tar.gz` and `.tar.xz`.
You can also use blank and it means binary is not compressed.

If you compress binary, uncompressed directory or file must contain the executable named `{cmd}`.

And you can also use `-` for separator instead of `_` if you like.

For example, if your command name is `foo-bar`, one of followings is expected to be put in release
page on GitHub as binary for platform `linux` and arch `amd64`.

- `foo-bar_linux_amd64` (executable)
- `foo-bar_linux_amd64.zip` (zip file)
- `foo-bar_linux_amd64.tar.gz` (tar file)
- `foo-bar_linux_amd64.xz` (xzip file)
- `foo-bar-linux-amd64.tar.gz` (`-` is also ok for separator)

If you compress and/or archive your release asset, it must contain an executable named one of followings:

- `foo-bar` (only command name)
- `foo-bar_linux_amd64` (full name)
- `foo-bar-linux-amd64` (`-` is also ok for separator)

To archive the executable directly on Windows, `.exe` can be added before file extension like
`foo-bar_windows_amd64.exe.zip`.

[goreleaser]: https://github.com/goreleaser/goreleaser/


### Naming Rules of Versions (=Git Tags)

go-selfupdate searches binaries' versions via Git tag names (not a release title).
When your tool's version is `1.2.3`, you should use the version number for tag of the Git
repository (i.e. `1.2.3` or `v1.2.3`).

This library assumes you adopt [semantic versioning][]. It is necessary for comparing versions
systematically.

Prefix before version number `\d+\.\d+\.\d+` is automatically omitted. For example, `ver1.2.3` or
`release-1.2.3` are also ok.

Tags which don't contain a version number are ignored (i.e. `nightly`). And releases marked as `pre-release`
are also ignored.

[semantic versioning]: https://semver.org/


### Structure of Releases

In summary, structure of releases on GitHub looks like:

- `v1.2.0`
  - `foo-bar-linux-amd64.tar.gz`
  - `foo-bar-linux-386.tar.gz`
  - `foo-bar-darwin-amd64.tar.gz`
  - `foo-bar-windows-amd64.zip`
  - ... (Other binaries for v1.2.0)
- `v1.1.3`
  - `foo-bar-linux-amd64.tar.gz`
  - `foo-bar-linux-386.tar.gz`
  - `foo-bar-darwin-amd64.tar.gz`
  - `foo-bar-windows-amd64.zip`
  - ... (Other binaries for v1.1.3)
- ... (older versions)

### Special case for ARM architecture

If you're using goreleaser targeting ARM CPUs, it will use the version of the ARM architecture as a name:
- `armv5`
- `armv6`
- `armv7`

go-selfupdate will check which architecture was used to build the current binary. Please note it's not detecting the hardware, but the binary target instead. If you run an `armv6` binary on an `armv7` CPU, it will keep `armv6` as a target.

As a rule, it will search for a binary with the same architecture first, then try the architectures below if available, and as a last resort will try a simple `arm` architecture tag.

So if you're running a `armv6` binary, it will try these targets in order:
- `armv6`
- `armv5`
- `arm`

More information on targeting ARM cpu can be found here: [GoArm](https://github.com/golang/go/wiki/GoArm)

### Hash or Signature Validation

go-selfupdate supports hash or signature validation of the downloaded files. It comes
with support for sha256 hashes or ECDSA signatures. In addition to internal functions the
user can implement the `Validator` interface for own validation mechanisms.

```go
// Validator represents an interface which enables additional validation of releases.
type Validator interface {
	// Validate validates release bytes against an additional asset bytes.
	// See SHAValidator or ECDSAValidator for more information.
	Validate(release, asset []byte) error
	// Suffix describes the additional file ending which is used for finding the
	// additional asset.
	Suffix() string
}
```

#### SHA256

To verify the integrity by SHA256 generate a hash sum and save it within a file which has the
same naming as original file with the suffix `.sha256`.
For e.g. use sha256sum, the file `selfupdate/testdata/foo.zip.sha256` is generated with:
```shell
sha256sum foo.zip > foo.zip.sha256
```

#### ECDSA
To verify the signature by ECDSA generate a signature and save it within a file which has the
same naming as original file with the suffix `.sig`.
For e.g. use openssl, the file `selfupdate/testdata/foo.zip.sig` is generated with:
```shell
openssl dgst -sha256 -sign Test.pem -out foo.zip.sig foo.zip
```

go-selfupdate makes use of go internal crypto package. Therefore the used private key
has to be compatible with FIPS 186-3.

### Using other providers than Github

This library can be easily extended by providing a new source and release implementation for any git provider
Currently implemented are 
- Github (default)
- Gitea 

Check the *-custom examples in cmd to see how a custom source like the GiteaSource can be used
### Copyright

This work is based on:


- [go-github-selfupdate](https://github.com/rhysd/go-github-selfupdate): [Copyright (c) 2017 rhysd](https://github.com/rhysd/go-github-selfupdate/blob/master/LICENSE)

- [go-update](https://github.com/inconshreveable/go-update): [Copyright 2015 Alan Shreve](https://github.com/inconshreveable/go-update/blob/master/LICENSE)
