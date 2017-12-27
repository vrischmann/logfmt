package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
)

type query struct {
	key    string
	value  string
	fuzzy  bool
	regexp *regexp.Regexp

	parser logfmt.PairParser
}

func (q *query) Match(line string) bool {
	pairs := q.parser.Split(line)

	var pair logfmt.Pair
	for _, v := range pairs {
		if v.Key == q.key {
			pair = v
			break
		}
	}

	if pair.Key == "" {
		return false
	}

	switch {
	case q.fuzzy:
		return strings.Contains(pair.Value, q.value)

	case q.regexp != nil:
		return q.regexp.MatchString(pair.Value)

	default:
		return pair.Value == q.value
	}
}

type queries []query

func (q queries) Match(or bool, line string) bool {
	if or {
		for _, qry := range q {
			if qry.Match(line) {
				return true
			}
		}
		return false
	}

	for _, qry := range q {
		if !qry.Match(line) {
			return false
		}
	}
	return true
}

const (
	regexOperator  = "=~"
	fuzzyOperator  = "~"
	strictOperator = "="
)

func extractQueries(args []string) queries {
	var res queries

	for _, arg := range args {
		switch {
		case strings.Contains(arg, regexOperator):
			tokens := strings.Split(arg, regexOperator)
			res = append(res, query{
				key:    tokens[0],
				regexp: regexp.MustCompile(tokens[1]),
			})

		case strings.Contains(arg, fuzzyOperator):
			tokens := strings.Split(arg, fuzzyOperator)
			res = append(res, query{
				key:   tokens[0],
				value: tokens[1],
				fuzzy: true,
			})

		case strings.Contains(arg, strictOperator):
			tokens := strings.Split(arg, strictOperator)
			res = append(res, query{
				key:   tokens[0],
				value: tokens[1],
				fuzzy: false,
			})

		default:
			return res
		}
	}

	return res
}

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
		flReverse = flag.Bool("v", false, "Reverse matches")
		flOr      = flag.Bool("or", false, "Treat multiple queries as a OR instead of an AND")
	)
	flag.Parse()

	args := flag.Args()

	queries := extractQueries(args)
	args = args[len(queries):]

	//

	input := internal.GetInput(args)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()

		matches := queries.Match(*flOr, line)
		switch {
		case matches && !*flReverse:
			io.WriteString(os.Stdout, line+"\n")
		case !matches && *flReverse:
			io.WriteString(os.Stdout, line+"\n")
		}
	}
}
