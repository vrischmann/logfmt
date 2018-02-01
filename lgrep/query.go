package lgrep

import (
	"regexp"
	"strings"

	"github.com/vrischmann/logfmt"
)

type Query struct {
	key    string
	value  string
	fuzzy  bool
	regexp *regexp.Regexp

	keyWithEquals string // used only in the fast failout
	parser        logfmt.PairParser
	pairs         logfmt.Pairs
}

func newQuery(key string) Query {
	return Query{
		key:           key,
		keyWithEquals: key + "=",
		pairs:         make(logfmt.Pairs, 64),
	}
}

func (q *Query) Copy() Query {
	tmp := Query{
		key:           q.key,
		keyWithEquals: q.keyWithEquals,
		value:         q.value,
		fuzzy:         q.fuzzy,
		pairs:         make(logfmt.Pairs, len(q.pairs)),
	}
	if q.regexp != nil {
		tmp.regexp = q.regexp.Copy()
	}
	return tmp
}

func (q *Query) MatchKeys(keys []string) bool {
	for _, key := range keys {
		if key == q.key {
			return true
		}
	}
	return false
}

func (q *Query) Match(line string) bool {
	// Fast bailout: if the key is not in the line there's no need to parse the line
	if !strings.Contains(line, q.keyWithEquals) {
		return false
	}

	pairs := q.parser.SplitInto(line, q.pairs)

	var pair *logfmt.Pair
	for i := range pairs {
		v := &pairs[i]

		if v.Key == q.key {
			pair = v
			break
		}
	}

	if pair == nil {
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

type Queries []Query

func (q Queries) Copy() Queries {
	tmp := make(Queries, len(q))
	for i, v := range q {
		tmp[i] = v.Copy()
	}
	return tmp
}

func (q Queries) MatchKeys(keys []string, opt *QueryOption) bool {
	if opt != nil && opt.Reverse {
		return !q.matchKeys(keys)
	}

	return q.matchKeys(keys)
}

func (q Queries) matchKeys(keys []string) bool {
	for i := range q {
		qry := &q[i]
		if !qry.MatchKeys(keys) {
			return false
		}
	}
	return true
}

type QueryOption struct {
	Reverse bool
	Or      bool
}

func (q Queries) match(line string, opt *QueryOption) bool {
	switch {
	case opt != nil && opt.Or:
		for i := range q {
			qry := &q[i]
			if qry.Match(line) {
				return true
			}
		}
		return false

	default:
		res := true
		for i := range q {
			qry := &q[i]
			if !qry.Match(line) {
				res = false
			}
		}
		return res
	}
}

func (q Queries) Match(line string, opt *QueryOption) bool {
	if opt != nil && opt.Reverse {
		return !q.match(line, opt)
	}

	return q.match(line, opt)
}

const (
	regexOperator  = "=~"
	fuzzyOperator  = "~"
	strictOperator = "="
)

func ExtractQueries(args []string) Queries {
	var res Queries

	for _, arg := range args {
		var qry Query

		switch {
		case strings.Contains(arg, regexOperator):
			tokens := strings.Split(arg, regexOperator)
			qry = newQuery(tokens[0])
			qry.regexp = regexp.MustCompile(tokens[1])

		case strings.Contains(arg, fuzzyOperator):
			tokens := strings.Split(arg, fuzzyOperator)
			qry = newQuery(tokens[0])
			qry.value = tokens[1]
			qry.fuzzy = true

		case strings.Contains(arg, strictOperator):
			tokens := strings.Split(arg, strictOperator)
			qry = newQuery(tokens[0])
			qry.value = tokens[1]

		default:
			return res
		}

		res = append(res, qry)
	}

	return res
}
