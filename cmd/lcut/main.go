package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
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

func main() {
	var (
		flReverse = flag.Bool("v", false, "Reverse matches")
		flInput   inputFiles
	)
	flag.Var(&flInput, "i", "The input files")

	flag.Parse()

	fields := cutFields(flag.Args())

	//

	input := internal.GetInput(flInput)

	var (
		buf     = make([]byte, 0, 4096)
		scanner = bufio.NewScanner(input)
	)
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
}
