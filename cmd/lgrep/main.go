package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"unsafe"

	"github.com/vrischmann/logfmt/internal"
	"github.com/vrischmann/logfmt/internal/flags"
	"github.com/vrischmann/logfmt/internal/lgrep"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of lgrep: lgrep [OPTION]... QUERY... [FILE]...")
		fmt.Fprintln(os.Stderr, "Search for QUERY in each FILE.")
		fmt.Fprintln(os.Stderr, "Multiple files are allowed. If no files, search from stdin.")
		fmt.Fprintln(os.Stderr, "QUERY must be in one of these form:")
		fmt.Fprintln(os.Stderr, "  city=Lyon                      for a strict match. Will only match lines which have the `city` key with the value Lyon.")
		fmt.Fprintln(os.Stderr, "  city~New                       for a fuzzy match. Will match lines which have the `city` key with any value contaning New.")
		fmt.Fprintln(os.Stderr, "  city=~(Paris|Lyon|San [a-z]+)  for a regexp match. Will match lines which have the `city` key and for which the regexp matches the value.")
		fmt.Fprintln(os.Stderr, "You can also trick lgrep to test for presence of a key by using a fuzzy match operator with no value to match:")
		fmt.Fprintln(os.Stderr, "  city~                          Will match lines which have the `city` key with any value (because any value contains the empty string).")
		fmt.Fprint(os.Stderr, "You can have multiple queries. By default it will work as an AND, you can treat them as a OR with the -or option.\n\n")
		fmt.Fprintln(os.Stderr, "Available options:")

		flag.PrintDefaults()
	}
}

func main() {
	var (
		flReverse      = flag.Bool("v", false, "Reverse matches")
		flWithFilename = flag.Bool("with-filename", false, "Display the filename")
		flOr           = flag.Bool("or", false, "Treat multiple queries as a OR instead of an AND")
	)
	flag.Parse()

	stopProfiling := internal.StartProfiling(flags.CPUProfile, flags.MemProfile)
	defer stopProfiling()

	//

	args := flag.Args()
	qs := lgrep.ExtractQueries(args)
	args = args[len(qs):]

	//

	inputs := internal.GetInputs(args)

	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		scanner.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))

		strHeader := new(reflect.StringHeader)

		for scanner.Scan() {
			data := scanner.Bytes()

			strHeader.Data = uintptr(unsafe.Pointer(&data[0]))
			strHeader.Len = len(data)

			line := *(*string)(unsafe.Pointer(strHeader))

			matches := qs.Match(*flOr, line)

			if (matches && !*flReverse) || (!matches && *flReverse) {
				printLine(*flWithFilename, input.Name, line)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}

func printLine(withFilename bool, inputName string, line string) {
	if withFilename {
		io.WriteString(os.Stdout, inputName+": "+line+"\n")
	} else {
		io.WriteString(os.Stdout, line+"\n")
	}
}
