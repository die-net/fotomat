package main

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Attempt to guess the number of physical (non-hyperthread) cores
// available.  If that doesn't work, return Go's count of virtual cores that
// we can be scheduled on.
func numCPUCores() int {
	ht := hyperthreadsPerCore()
	switch {
	case ht > 0:
		// Valid value: We are on Linux and /proc/cpuinfo was available.
	case runtime.GOARCH == "amd64":
		ht = 2 // Assume amd64 has 2 virtual cores per physical core.
	default:
		ht = 1 // Otherwise, assume no hyperthreading.
	}

	// runtime.NumCPU() uses sched_getaffinity to get the number of CPUs
	// we are allowed to be scheduled on.  Divide that by hyperthreads
	// per core and round up.
	return (runtime.NumCPU() + ht - 1) / ht
}

// On Linux x86, return count of hyperthreads per CPU core or 0 on error.
func hyperthreadsPerCore() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	// Parse cpuinfo looking for the highest "siblings" and "cpu cores"
	// values.
	siblings := 0
	cores := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f := strings.SplitN(scanner.Text(), ":", 2)
		if len(f) != 2 {
			continue
		}

		value, _ := strconv.Atoi(strings.TrimSpace(f[1]))
		switch strings.TrimSpace(f[0]) {
		case "siblings":
			if value > siblings {
				siblings = value
			}
		case "cpu cores":
			if value > cores {
				cores = value
			}
		}
	}

	// If both values seem reasonable, return ratio.
	if cores > 0 && siblings >= cores {
		return siblings / cores
	}

	return 0
}
