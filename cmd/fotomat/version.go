package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	// Updated by git-hooks/pre-commit
	FotomatVersion = "2.0.143"
)

var (
	version = flag.Bool("version", false, "Show version and exit.")
)

func showVersion() {
	fmt.Println("Fotomat v" + FotomatVersion)
	os.Exit(0)
}
