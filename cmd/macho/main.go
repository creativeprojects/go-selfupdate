// Macho reports whether its own executable is a macOS universal binary.
//
// The Makefile builds it for amd64 and arm64 and merges both with lipo
// into a universal binary.
package main

import (
	"debug/macho"
	"fmt"
	"os"
)

func main() {
	fatFile, err := macho.OpenFat(os.Args[0])
	if err != nil {
		fmt.Printf("not a universal binary: %s\n", err)
	} else {
		fmt.Printf("this is a universal binary\n")
		fatFile.Close()
	}
}
