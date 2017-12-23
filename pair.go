package logfmt

import (
	"strconv"
	"strings"
)

type Pair struct {
	Key   string
	Value string
}

type Pairs []Pair

func (p Pairs) Len() int {
	return len(p)
}

func (p Pairs) Less(i, j int) bool {
	return p[i].Key < p[j].Key
}

func (p Pairs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Pairs) AppendFormat(b []byte) []byte {
	for i, pair := range p {
		b = append(b, pair.Key...)
		b = append(b, '=')

		if needsQuoting(pair.Value) {
			b = append(b, strconv.Quote(pair.Value)...)
		} else {
			b = append(b, pair.Value...)
		}

		if i+1 < len(p) {
			b = append(b, ' ')
		}
	}

	return b
}

func needsQuoting(s string) bool {
	return strings.ContainsAny(s, " \"")
}

func (p Pairs) Format() string {
	return string(p.AppendFormat(nil))
}
