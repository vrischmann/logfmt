package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/cobra"

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

func (i inputFiles) Type() string { return "string" }

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

func runMain(cmd *cobra.Command, args []string) error {
	stopProfiling := internal.StartProfiling(flags.CPUProfile, flags.MemProfile)
	defer stopProfiling()

	//

	fields := cutFields(args)
	inputs := internal.GetInputs(flInput)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		scanner.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			pairs = fields.CutFrom(flReverse, pairs)

			if len(pairs) <= 0 {
				continue
			}

			buf = pairs.AppendFormat(buf)
			buf = append(buf, '\n')

			_, err := os.Stdout.Write(buf)
			if err != nil {
				return err
			}

			buf = buf[:0]
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
	flReverse bool
	flInput   inputFiles

	rootCmd = &cobra.Command{
		Use:   "lcut [field]",
		Short: `cut "field" from each input log line`,
		Long: `cut "field" from each input log line

Multiple fields are allowed. Does nothing if no fields are specified.`,
		Args: cobra.MinimumNArgs(1),
		RunE: runMain,
	}
)

func init() {
	fs := rootCmd.Flags()

	fs.BoolVarP(&flReverse, "reverse", "v", false, "Reverse cut: keep only the fields provided")
	fs.VarP(&flInput, "input", "i", "Use these input files instead of stdin")
	fs.Var(&flags.MaxLineSize, "max-line-size", "Max size in bytes of a line")
	fs.StringVar(&flags.CPUProfile, "cpu-profile", "", "Writes a CPU profile at `cpu-profile` after execution")
	fs.StringVar(&flags.MemProfile, "mem-profile", "", "Writes a memory profile at `mem-profile` after execution")
}
