package main

import (
	"bufio"
	"os"
	"runtime"
	"strings"
)

// Attempt to guess the number of physical (non-hyperthread) cores
// available.  If that doesn't work, return Go's count of virtual cores that
// we can be scheduled on.
func numCpuCores() int {
	cpus := runtime.NumCPU()

	cores := procCpuinfoCores()
	if cores > 0 && cores <= cpus {
		return cores
	}

	// Assume non-Linux amd64 has 2 virtual cores per physical core.
	if runtime.GOARCH == "amd64" {
		cores = (cores + 1) / 2
	}

	return cpus
}

// On Linux x86, count unique physical id + core id pairs, which should give
// total non-Hyperthreaded cores available.  Return 0 on error.
func procCpuinfoCores() int {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	phys := ""
	s := make(map[string]bool)
	for scanner.Scan() {
		f := strings.SplitN(scanner.Text(), ":", 2)
		if len(f) != 2 {
			continue
		}

		value := strings.TrimSpace(f[1])
		switch strings.TrimSpace(f[0]) {
		case "physical id":
			phys = value
		case "core id":
			if phys != "" {
				s[phys+"-"+value] = true
			}
		}
	}
	return len(s)
}
