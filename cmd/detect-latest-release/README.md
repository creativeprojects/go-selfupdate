This command line tool is a small wrapper of [`Updater.DetectLatest()`](https://pkg.go.dev/github.com/creativeprojects/go-selfupdate/v2#Updater.DetectLatest).

Please install using `go install`.

```
$ go install github.com/creativeprojects/go-selfupdate/v2/cmd/detect-latest-release@latest
```

To know the usage, please try the command without any argument.

```
$ detect-latest-release
```

For example, following shows the latest version of [resticprofile](https://github.com/creativeprojects/resticprofile).

```
$ detect-latest-release creativeprojects/resticprofile
```

The repository can be an `owner/name` slug or the URL of the repository. Flags:

- `-t`: version control, one of `github`, `gitea`, `gitlab` or `http`. The default `auto` picks `gitea` or `gitlab` when the domain name contains it, and `github` otherwise.
- `-u`: base URL of the server when the repository is a plain `owner/name` slug, for dedicated instances and HTTP repositories.
- `-o`, `-a`: OS and CPU architecture to look for, instead of the ones the tool was built for.
- `-v`: display debugging information.
