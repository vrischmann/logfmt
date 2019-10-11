package internal

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func StartProfiling(cpuProfile, memProfile string) func() {
	if cpuProfile == "" && memProfile == "" {
		return func() {}
	}

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
	}

	if cpuProfile != "" && memProfile == "" {
		return pprof.StopCPUProfile
	}

	return func() {
		pprof.StopCPUProfile()

		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal(err)
		}

		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}

		f.Close()
	}
}
