Self-Update library for GitHub hosted applications in Go
========================================================

[![Godoc reference](https://godoc.org/github.com/creativeprojects/go-github-selfupdate?status.svg)](http://godoc.org/github.com/creativeprojects/go-github-selfupdate)
[![Build Status](https://travis-ci.com/creativeprojects/go-github-selfupdate.svg?branch=main)](https://travis-ci.com/creativeprojects/go-github-selfupdate)
[![codecov](https://codecov.io/gh/creativeprojects/go-github-selfupdate/branch/main/graph/badge.svg?token=3FejM0fkw2)](https://codecov.io/gh/creativeprojects/go-github-selfupdate)

go-github-selfupdate detects the information of the latest release via [GitHub Releases API][] and
checks the current version. If a newer version than itself is detected, it downloads the released binary from
GitHub and replaces itself.

- Automatically detect the latest version of released binary on GitHub
- Retrieve the proper binary for the OS and arch where the binary is running
- Update the binary with rollback support on failure
- Tested on Linux, macOS and Windows
- Many archive and compression formats are supported (zip, tar, gzip, xzip)
- Support private repositories
- Support hash, signature validation

[GitHub Releases API]: https://developer.github.com/v3/repos/releases/

