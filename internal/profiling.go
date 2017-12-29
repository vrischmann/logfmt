package internal

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
)

func StartProfiling(cpuProfile, memProfile string) func() {
	if cpuProfile == "" && memProfile == "" {
		return func() {}
	}

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			logrus.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			logrus.Fatal(err)
		}
	}

	if cpuProfile != "" && memProfile == "" {
		return pprof.StopCPUProfile
	}

	return func() {
		f, err := os.Create(memProfile)
		if err != nil {
			logrus.Fatal(err)
		}

		runtime.GC()

		if err := pprof.WriteHeapProfile(f); err != nil {
			logrus.Fatal(err)
		}

		f.Close()
	}
}
