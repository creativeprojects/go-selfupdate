This command line tool is a small wrapper of [`selfupdate.DetectLatest()`](https://pkg.go.dev/github.com/creativeprojects/go-selfupdate/selfupdate#DetectLatest).

Please install using `go get`.

```
$ go get -u github.com/creativeprojects/go-selfupdate/cmd/detect-latest-release
```

To know the usage, please try the command without any argument.

```
$ detect-latest-release
```

For example, following shows the latest version of [github-clone-all](https://github.com/creativeprojects/resticprofile).

```
$ detect-latest-release creativeprojects/resticprofile
```

