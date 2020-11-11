Self-Update library for GitHub hosted applications in Go
========================================================

[![Godoc reference](https://godoc.org/github.com/creativeprojects/go-selfupdate?status.svg)](http://godoc.org/github.com/creativeprojects/go-selfupdate)
[![Build Status](https://travis-ci.com/creativeprojects/go-selfupdate.svg?branch=main)](https://travis-ci.com/creativeprojects/go-selfupdate)
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
