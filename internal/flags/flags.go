package flags

import "flag"

var (
	CPUProfile  string
	MemProfile  string
	MaxLineSize int
)

func init() {
	flag.IntVar(&MaxLineSize, "max-line-size", 1024*1024, "Max size in bytes of a line (default %d)")

	flag.StringVar(&CPUProfile, "cpu-profile", "", "Writes a CPU profile at `cpu-profile` after execution")
	flag.StringVar(&MemProfile, "mem-profile", "", "Writes a memory profile at `mem-profile` after execution")
}
