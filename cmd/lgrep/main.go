package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vrischmann/logfmt"
)

type query struct {
	key    string
	value  string
	fuzzy  bool
	regexp *regexp.Regexp
}

func (q *query) Match(line string) bool {
	pairs := logfmt.Split(line)

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

func (q queries) Match(line string) bool {
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

func main() {
	var (
		flReverse = flag.Bool("v", false, "Reverse matches")
	)
	flag.Parse()

	args := flag.Args()

	queries := extractQueries(args)
	args = args[len(queries):]

	// determine the input data

	var input io.Reader = os.Stdin
	if len(args) > 0 {
		files := make([]io.Reader, 0, flag.NArg())
		for _, fileName := range args {
			f, err := os.Open(fileName)
			if err != nil {
				logrus.Fatal(err)
			}

			files = append(files, f)
		}

		input = io.MultiReader(files...)
	}

	//

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()

		matches := queries.Match(line)
		switch {
		case matches && !*flReverse:
			io.WriteString(os.Stdout, line)
		case !matches && *flReverse:
			io.WriteString(os.Stdout, line)
		}
	}
}
