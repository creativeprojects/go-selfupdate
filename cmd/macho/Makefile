.PHONY: build
build:
	GOOS=darwin GOARCH=amd64 go build -o macho_amd64 .
	GOOS=darwin GOARCH=arm64 go build -o macho_arm64 .
	lipo -create -output macho_universal macho_amd64 macho_arm64
