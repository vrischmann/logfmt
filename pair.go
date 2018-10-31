package logfmt

import (
	"strconv"
	"strings"
)

// Pair contains a key and value of a logfmt line.
type Pair struct {
	Key   string
	Value string
}

// Pairs is a collection of key-value pairs.
// Pairs implements sort.Interface so it is sortable.
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

// AppendFormat formats the pairs in a logfmt compatible way and appends the formatted strings to b.
// It returns the resulting b.
//
// Note that the pairs are appended as they come, there's no reordering.
// If you want the pairs to be sorted you have to call `sort.Sort` on the slice first.
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
	return strings.ContainsAny(s, " \"=")
}

// Format formats the pairs in a logfmt compatible way.
func (p Pairs) Format() string {
	return string(p.AppendFormat(nil))
}
