package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vrischmann/logfmt/internal"
	"github.com/vrischmann/logfmt/internal/flags"
	"github.com/vrischmann/logfmt/lgrep"
)

var newLineBytes = []byte{'\n'}

func runMain(cmd *cobra.Command, args []string) error {
	stopProfiling := internal.StartProfiling(flags.CPUProfile, flags.MemProfile)
	defer stopProfiling()

	//

	qs := lgrep.ExtractQueries(args)
	args = args[len(qs):]

	//

	inputs := internal.GetInputs(args)

	qryOpt := &lgrep.QueryOption{
		Or:      flOr,
		Reverse: flReverse,
	}

	for _, input := range inputs {
		var scanner scanner
		switch {
		case input.Reader != nil:
			scanner = newStdScanner(input.Reader)
		case input.Data != nil:
			scanner = newNoCopyScanner(input.Data)
		}

		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) <= 0 {
				continue
			}

			line := scanner.String()

			if qs.Match(line, qryOpt) {
				if flWithFilename {
					io.WriteString(os.Stdout, input.Name+": "+line+"\n")
				} else {
					os.Stdout.Write(data)
					os.Stdout.Write(newLineBytes)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	rootCmd.Execute()
}

var (
	rootCmd = &cobra.Command{
		Use:   "lgrep [query] [file]",
		Short: `search for "query" in each "file"`,
		Long: `search for "query" in each "file".

Multiple files are allowed. If no files, search from stdin.

QUERY must be in one of these form:

    city=Lyon                      for a strict match. Will only match lines which have the "city" key with the value Lyon.
    city~New                       for a fuzzy match. Will match lines which have the "city" key with any value contaning New.
    city=~(Paris|Lyon|San [a-z]+)  for a regexp match. Will match lines which have the "city" key and for which the regexp matches the value.

You can also trick lgrep to test for presence of a key by using a fuzzy match operator with no value to match:
    city~                          Will match lines which have the "city" key with any value (because any value contains the empty string).

You can have multiple queries. By default it will work as an AND, you can treat them as a OR with the --or option.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runMain,
	}

	flReverse      bool
	flWithFilename bool
	flOr           bool
)

func init() {
	fs := rootCmd.Flags()

	fs.BoolVarP(&flReverse, "reverse", "v", false, "Reverse matches")
	fs.BoolVarP(&flWithFilename, "with-filename", "H", false, "Display the filename")
	fs.BoolVarP(&flOr, "or", "o", false, "Treat multiple queries as a OR instead of a AND")
	fs.Var(&flags.MaxLineSize, "max-line-size", "Max size in bytes of a line (default %d)")
	fs.StringVar(&flags.CPUProfile, "cpu-profile", "", "Writes a CPU profile at `cpu-profile` after execution")
	fs.StringVar(&flags.MemProfile, "mem-profile", "", "Writes a memory profile at `mem-profile` after execution")
}
