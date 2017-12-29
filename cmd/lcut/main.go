package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
	"github.com/vrischmann/logfmt/internal/flags"
)

type inputFiles []string

func (i *inputFiles) Set(s string) error {
	*i = append(*i, s)
	return nil
}

func (i *inputFiles) String() string {
	return strings.Join(*i, ",")
}

type cutFields []string

func (f cutFields) CutFrom(reverse bool, pairs logfmt.Pairs) logfmt.Pairs {
	if len(f) == 0 {
		return pairs
	}

	res := pairs[:0]
	for _, pair := range pairs {
		for _, field := range f {
			ok := !reverse && pair.Key != field
			reverseOK := reverse && pair.Key == field

			if ok || reverseOK {
				res = append(res, pair)
			}
		}
	}

	return res
}

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of lcut: lcut [OPTION]... FIELD...")
		fmt.Fprintln(os.Stderr, "Cut FIELD from each line of the inputs")
		fmt.Fprintln(os.Stderr, "Multiple fields are allowed. Does nothing if no fields are specified.")
		fmt.Fprintln(os.Stderr, "Available options:")
		flag.PrintDefaults()
	}
}

func main() {
	var (
		flReverse = flag.Bool("v", false, "Reverse cut: keep only the fields provided")
		flInput   inputFiles
	)
	flag.Var(&flInput, "i", "The input files. Can be specified multiple times. If no files, read from stdin")

	flag.Parse()

	stopProfiling := internal.StartProfiling(flags.CPUProfile, flags.MemProfile)
	defer stopProfiling()

	//

	fields := cutFields(flag.Args())

	//

	inputs := internal.GetInputs(flInput)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		scanner.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			pairs = fields.CutFrom(*flReverse, pairs)

			if len(pairs) <= 0 {
				continue
			}

			buf = pairs.AppendFormat(buf)
			buf = append(buf, '\n')

			_, err := os.Stdout.Write(buf)
			if err != nil {
				log.Fatal(err)
			}

			buf = buf[:0]
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}
