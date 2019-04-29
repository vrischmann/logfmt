package main

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
	"github.com/vrischmann/logfmt/internal/flags"
)

type sortElement struct {
	line  string
	field string
}

type sortAlphabetical []sortElement

func (s sortAlphabetical) Len() int           { return len(s) }
func (s sortAlphabetical) Less(i, j int) bool { return s[i].field < s[j].field }
func (s sortAlphabetical) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type sortNumerical []sortElement

func (s sortNumerical) Len() int { return len(s) }
func (s sortNumerical) Less(i, j int) bool {
	a, err := strconv.ParseInt(s[i].field, 10, 64)
	if err != nil {
		logrus.Fatalf("field %q is invalid for a numerical sort", s[i].field)
	}

	b, err := strconv.ParseInt(s[j].field, 10, 64)
	if err != nil {
		logrus.Fatalf("field %q is invalid for a numerical sort", s[j].field)
	}

	return a < b
}
func (s sortNumerical) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortByDuration []sortElement

func (s sortByDuration) Len() int { return len(s) }
func (s sortByDuration) Less(i, j int) bool {
	a, err := time.ParseDuration(s[i].field)
	if err != nil {
		logrus.Fatalf("field %q is invalid for a duration sort", s[i].field)
	}

	b, err := time.ParseDuration(s[j].field)
	if err != nil {
		logrus.Fatalf("field %q is invalid for a duration sort", s[j].field)
	}

	return a < b
}
func (s sortByDuration) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func getFromPairs(pairs logfmt.Pairs, key string) string {
	for _, pair := range pairs {
		if pair.Key == key {
			return pair.Value
		}
	}
	return ""
}

func runMain(cmd *cobra.Command, args []string) error {
	stopProfiling := internal.StartProfiling(flags.CPUProfile, flags.MemProfile)
	defer stopProfiling()

	//

	inputs := internal.GetInputs(nil)
	field := args[0]

	lines := make([]sortElement, 0, 8192)

	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		scanner.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			if len(pairs) <= 0 {
				continue
			}

			val := getFromPairs(pairs, field)
			if val == "" {
				continue
			}

			lines = append(lines, sortElement{
				line:  line,
				field: val,
			})
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	//

	var sl sort.Interface
	switch {
	case flNumericSort:
		sl = sortNumerical(lines)
	case flDurationSort:
		sl = sortByDuration(lines)
	default:
		sl = sortAlphabetical(lines)
	}

	switch {
	case flReverse:
		sort.Sort(sort.Reverse(sl))
	default:
		sort.Sort(sl)
	}

	//

	for _, line := range lines {
		os.Stdout.Write([]byte(line.line + "\n"))
	}

	return nil
}

func main() {
	rootCmd.Execute()
}

var (
	flReverse      bool
	flNumericSort  bool
	flDurationSort bool

	rootCmd = &cobra.Command{
		Use:   "lsort [field]",
		Short: `sort all input based on "field"`,
		Long: `sort all input based on "field"

Only one field is allowed.

Sorting can be alphabetical (the default), numeric (with -n) or by duration (with -d). For example:

    $ cat foobar.txt
	id=200 bar=baz
	id=3 baz=qux
	id=10 foo=bar
	$ cat foobar.txt | lsort id
	id=10 foo=bar
	id=200 bar=baz
	id=3 baz=qux

Numeric sort:

    $ cat foobar.txt
	id=200 bar=baz
	id=3 baz=qux
	id=10 foo=bar
	$ cat foobar.txt | lsort -n id
	id=3 baz=qux
	id=10 foo=bar
	id=200 bar=baz

Duration sort:

    $ cat foobar.txt
	elapsed=10m22s foo=bar
	elapsed=3m bar=baz
	elapsed=12s baz=qux
	$ cat foobar.txt | lsort -d id
	elapsed=12s baz=qux
	elapsed=3m bar=baz
	elapsed=10m22s foo=bar

Note: sorting is done in memory for now so be careful with your input data.`,
		Args: cobra.ExactArgs(1),
		RunE: runMain,
	}
)

func init() {
	fs := rootCmd.Flags()

	fs.BoolVarP(&flReverse, "reverse", "v", false, "Reverse sort")
	fs.BoolVarP(&flNumericSort, "numeric-sort", "n", false, "Use a numeric sort instead of a alphabetical sort")
	fs.BoolVarP(&flDurationSort, "duration-sort", "d", false, "Use a duration sort instead of a alphabetical sort")
	fs.Var(&flags.MaxLineSize, "max-line-size", "Max size in bytes of a line")
	fs.StringVar(&flags.CPUProfile, "cpu-profile", "", "Writes a CPU profile at `cpu-profile` after execution")
	fs.StringVar(&flags.MemProfile, "mem-profile", "", "Writes a memory profile at `mem-profile` after execution")
}
