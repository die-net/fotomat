package main

import (
	"flag"
	"log"
	"os"
	"path"
	"strconv"
)

var (
	tempDir = flag.String("temp_directory", os.TempDir(), "Path to store temporary files.")
)

func setupTempdir() {
	// Make sure we don't accidentally destroy things if someone passes
	// in / or ~ as tempDir.  Include uid in path to avoid permissions
	// problems.
	*tempDir = path.Join(*tempDir, "fotomat_temp-"+strconv.Itoa(os.Getuid()))

	if err := os.RemoveAll(*tempDir); err != nil {
		log.Fatalln("Can't remove directory", *tempDir, err)
	}

	if err := os.MkdirAll(*tempDir, 0700); err != nil {
		log.Fatalln("Can't create directory", *tempDir, err)
	}
}
