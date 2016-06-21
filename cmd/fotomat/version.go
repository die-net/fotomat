package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	// FotomatVersion is updated by git-hooks/pre-commit
	FotomatVersion = "2.4.182"
)

var (
	version = flag.Bool("version", false, "Show version and exit.")
)

func showVersion() {
	if *version {
		fmt.Println("Fotomat v" + FotomatVersion)
		os.Exit(0)
	}
}

func init() {
	post(showVersion)
}
