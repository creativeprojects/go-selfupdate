# Contributing to go-selfupdate

Thank you for taking the time to contribute. This document explains how to set up a development environment, how the tests and linters work, and what to expect when you open a pull request.

## Table of contents

- [Reporting issues](#reporting-issues)
- [Before you start a change](#before-you-start-a-change)
- [Development setup](#development-setup)
- [Running the tests](#running-the-tests)
- [Linting](#linting)
- [Code guidelines](#code-guidelines)
- [Writing tests](#writing-tests)
- [Adding a new source provider](#adding-a-new-source-provider)
- [Adding a new validator](#adding-a-new-validator)
- [Documentation](#documentation)
- [Dependencies](#dependencies)
- [Submitting a pull request](#submitting-a-pull-request)
- [Releases](#releases)
- [License](#license)

## Reporting issues

Bug reports and feature requests go in the [GitHub issue tracker](https://github.com/creativeprojects/go-selfupdate/issues).

A useful bug report includes:

- the version of go-selfupdate you are using (the tag or commit hash)
- the Go version and the OS/architecture the binary is running on
- which source you use (GitHub, GitHub Enterprise, GitLab, Gitea or HTTP)
- the names of the release assets, and the tag name of the release, when the problem is about detecting or picking an asset
- what you expected to happen, and what happened instead

**Security issues**: please do not open a public issue for a vulnerability. Use GitHub's private vulnerability reporting on the repository instead.

## Before you start a change

- For bug fixes and small improvements, go straight to a pull request.
- For new features, new providers, or anything that changes the public API, open an issue first to discuss the design. This avoids spending time on something that cannot be merged.
- Check the open issues and pull requests to make sure nobody is already working on the same thing.

## Development setup

You need Go 1.26 or later. The CI pipeline runs on the two most recent Go versions declared in [build.yml](.github/workflows/build.yml).

```sh
git clone https://github.com/creativeprojects/go-selfupdate.git
cd go-selfupdate
go mod download
make build
```

`make build` compiles every package, including the command line tools under `cmd/`.

The repository contains these packages:

| Package | Purpose |
|---|---|
| `.` (root, `selfupdate`) | The library: release detection, asset selection, download, decompression, validation and update |
| `github/`, `gitlab/`, `gitea/`, `http/` | One `selfupdate.Source` implementation per provider |
| `pgp/` | PGP signature validator, in its own package so the OpenPGP library is only linked when used |
| `update/` | Low level binary replacement with rollback, derived from `inconshreveable/go-update` |
| `internal/` | Executable path resolution and API token helpers, platform specific |
| `internal/memory` | In-memory `selfupdate.Source` used by the tests |
| `internal/sourcetest` | Conformance test suite for sources, with fake GitHub, GitLab, Gitea and HTTP servers |
| `cmd/` | Shared helpers for the command line tools |
| `cmd/detect-latest-release` | Tool that prints the latest release of a repository |
| `cmd/get-release` | Tool that downloads a release binary for the current platform |
| `cmd/serve-repo` | Minimal HTTP server to test the HTTP source locally |
| `cmd/macho` | Tool that tells whether a macOS binary is a universal binary |

## Running the tests

```sh
make test
```

This runs `go test -race -v ./...`, every package including `cmd/` and `internal/`. The CI pipeline runs the same command, adding `-short` when no `GITHUB_TOKEN` is set.

### Tests that reach the network

Most tests are offline: they use the in-memory source from `internal/memory`, populated by `mockMemorySource(t)` in [mock_data_test.go](mock_data_test.go), and the fixtures in [testdata/](testdata/). Each provider package runs the conformance suite from `internal/sourcetest` against a fake server started with `httptest`, so the providers are tested offline too.

Some tests talk to real servers. They are controlled in two ways:

- **`-short` flag**: tests that query the public GitHub API for the `creativeprojects/resticprofile` repository only run when `-short` is **not** set. Use `go test -short ./...` to run a fully offline test suite.
- **Environment variables**: the tests creating a source with a real token, in `github/source_test.go`, `gitlab/source_test.go` and `gitea/source_test.go`, are skipped unless `GITHUB_TOKEN`, `GITLAB_TOKEN` or `GITEA_TOKEN` respectively is set.

When running the GitHub tests without a token you may hit the API rate limit. Those tests detect the rate limit error and skip themselves rather than fail. Setting `GITHUB_TOKEN` to a personal access token (no scopes needed) avoids this.

The CI pipeline is written to run the network tests when a `GITHUB_TOKEN` environment variable is present and to fall back to `-short` otherwise. No token is currently exported to the test step, so CI runs the offline suite only. If your change touches a provider, run its tests locally with the matching token before opening the pull request, and say so in the description.

### Platform specific tests

The CI matrix runs the tests on Linux, macOS and Windows. A few tests are skipped on Windows, for example those creating symlinks, because they need administrator privileges there. Mirror this pattern with `t.Skip` and a reason if you need to skip a test on a platform; do not use build tags to hide tests.

### Coverage

```sh
make coverage        # offline tests, opens the HTML report
make full-coverage   # all tests, opens the HTML report
```

Coverage is uploaded to Codecov from the CI pipeline. The configuration in [codecov.yml](codecov.yml) expects new code in a pull request to be at least 70% covered. This is a target, not a hard gate, but please add tests with your change.

## Linting

The project uses [golangci-lint](https://golangci-lint.run/) version 2, with the configuration in [.golangci.yml](.golangci.yml).

```sh
make lint
```

The first run installs [eget](https://github.com/zyedidia/eget) and then the pinned version of golangci-lint into your `GOBIN` as `golangci-lint-v2`. The target runs the linter three times with `GOOS` set to `darwin`, `linux` and `windows`, because some files only compile on one platform. The CI pipeline runs the same version of golangci-lint, so a clean `make lint` locally means a clean lint job in CI.

To apply the automatic fixes, tidy the module files and run `go fix`:

```sh
make fix
```

The lint configuration does not enable a formatter, so run `gofmt` yourself or let your editor do it on save. Unformatted code will be asked to be fixed in review.

## Code guidelines

- **Cross platform first.** Everything must build and pass on Linux, macOS and Windows. When behaviour differs, put it in a file with a platform suffix (`hide_windows.go`, `resolve_path_unix.go`) and provide a no-op or fallback for the other platforms (`hide_noop.go`).
- **Keep the public API stable.** This library is imported by many projects. Do not rename or remove exported identifiers, change function signatures or change the semantics of existing options without discussing it in an issue first. Prefer adding a new option or function.
- **Pass a `context.Context`** to anything that does network or file I/O, as the `Source` interface does. Do not store contexts in structs.
- **Return sentinel errors** from [errors.go](errors.go) where one exists, and wrap them with `fmt.Errorf("...: %w", err)` so callers can use `errors.Is`. Add a new `Err...` variable there when a new error condition is worth distinguishing.
- **Log through the configured logger.** Use the `logger` field of the `Updater` or of the `Source` (see [log.go](log.go)) rather than `fmt.Print` or the standard `log` package. The library is silent by default, and users opt in by setting `Logger` in the `Config` of the updater and of the source.
- **Document exported identifiers** with a Go doc comment. The comment starts with the identifier name.
- **Keep comments short** and only where the code is not obvious on its own.
- **Avoid new dependencies.** See [Dependencies](#dependencies).

## Writing tests

- Tests use [testify](https://github.com/stretchr/testify). Use `require` for preconditions that make the rest of the test meaningless when they fail, and `assert` for the actual checks.
- Prefer table driven tests. Most existing tests follow this shape, with a `run bool` field for cases that depend on `testing.Short()`.
- Use `mockMemorySource(t)` from [mock_data_test.go](mock_data_test.go) when you need a source populated with releases and assets, and `mockSignedSourceRepository(t)` for one with signed checksum files. Extend the mock data if your test needs a new asset shape, and keep the existing cases passing.
- Put binary fixtures in [testdata/](testdata/). Keep them as small as possible, and add a note in the pull request on how they were produced if it is not obvious.
- Use `t.TempDir()` for anything written to disk.
- Use `t.Parallel()` when the test does not depend on global state (environment variables, the working directory).
- Do not hard code paths with `/` or `\`. Use `filepath.Join`.

## Adding a new source provider

A provider is a package implementing the three interfaces in [source.go](source.go): `Source`, `SourceRelease` and `SourceAsset`. The `gitea` package is the simplest to use as a template:

1. Create a package `<name>/` with a `source.go` holding a `Config` struct, a `NewSource` constructor and the `ListReleases`, `GetLatestRelease` and `DownloadReleaseAsset` methods. Read an API token from a `<NAME>_TOKEN` environment variable when none is given in the config, like the existing providers. Return `selfupdate.ErrRepositoryNotFound` and `selfupdate.ErrReleaseNotFound` where the `Source` interface documents them.
2. Add a `release.go` with the `Release` and `Asset` types implementing `SourceRelease` and `SourceAsset`.
3. Add a `source_test.go` running the conformance suite with `sourcetest.Run` against a fake server. Add the fake server to [internal/sourcetest/](internal/sourcetest/) when the provider API is new, and list the release fields the provider cannot report in `sourcetest.Options`. Tests needing a real token must skip when the environment variable is not set.
4. Register the provider in [cmd/source.go](cmd/source.go) so the command line tools can use it with the `-t` flag.
5. Add a section to the README describing how to use it, and update the title line, the `Introduction` and the `Providers` sections. Add the package to the list in [doc.go](doc.go).

## Adding a new validator

A validator is an implementation of the `Validator` interface in [validate.go](validate.go): a `Validate` method that checks an asset against a validation file, and a `GetValidationAssetName` method that returns the name of that file from the name of the asset. Look at `SHAValidator` for a minimal example and at `ChecksumValidator` for one that reads a file listing several assets.

Add the tests to [validate_test.go](validate_test.go), with any signature or key fixtures in [testdata/](testdata/), and document the new validator in the `Hash or Signature Validation` section of the README.

A validator pulling in a large dependency goes in its own package, like `pgp/`, so applications not using it do not link the library.

## Documentation

The README is the user facing documentation. If you change behaviour, naming rules, or add an option, update the README in the same pull request. The package comments in [doc.go](doc.go) and in each sub-package are published on pkg.go.dev: keep them in step with the README.

The table of contents at the top of the README is generated. After adding or renaming a section, regenerate it with:

```sh
make toc
```

This installs `gh-md-toc` and rewrites the block between `<!--ts-->` and `<!--te-->`. The target uses BSD `sed` syntax and is only known to work on macOS. On Linux, regenerate the table of contents by hand or edit the Makefile locally and do not commit that change.

Each command line tool under `cmd/` has its own README.

## Dependencies

This is a library, so every dependency it pulls in ends up in the build of every program that uses it. Please:

- prefer the standard library,
- do not add a dependency for something that can be done in a few lines,
- do not upgrade dependencies unrelated to your change, the maintainer does this in dedicated commits,
- run `go mod tidy` before committing so `go.mod` and `go.sum` stay minimal.

Dependency upgrades that fix a published vulnerability are always welcome as a standalone pull request. Mention the CVE in the commit message.

## Submitting a pull request

1. Fork the repository and create a branch from `main`.
2. Make your change. Keep each pull request to one logical change, it makes review faster.
3. Add or update tests, and update the documentation.
4. Run the full check locally:

   ```sh
   make build test lint
   ```

5. Push the branch and open a pull request against `main`. In the description, explain what the change does and why, link the related issue if there is one, and say how you tested it. Mention anything the reviewer should look at closely.
6. Make sure the CI pipeline is green on all six combinations of OS and Go version.

Commit messages are short and written in the imperative mood ("add support for ...", "fix ..."). A `fix:` or `chore:` prefix is welcome but not required. Keep the first line under 72 characters.

Do not commit coverage files, binaries built from `cmd/`, or editor configuration. The `.gitignore` already covers the usual ones.

Reviews can take a little time, as this project is maintained in spare time. Feel free to ping the pull request if you have not heard back in a couple of weeks.

## Releases

Releases are tagged by the maintainer with a semantic version (`v2.x.y`). Contributors do not need to bump any version number. A bug fix or a backward compatible addition goes in the next minor or patch release, a breaking change would need a new major version and is unlikely to be accepted without a very good reason.

## License

The project is released under the [MIT License](LICENSE). The `update/` package is derived from `inconshreveable/go-update` and carries its own [Apache 2.0 license](update/LICENSE).

By contributing, you agree that your contributions are licensed under the same license as the code they modify.
