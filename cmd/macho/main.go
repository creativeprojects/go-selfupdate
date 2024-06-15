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
