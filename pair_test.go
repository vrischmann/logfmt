package logfmt

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPairsSort(t *testing.T) {
	testCases := []struct {
		input Pairs
		exp   Pairs
	}{
		{
			Pairs{
				{"foo", "a b c d e f"},
				{"bar", "baz"},
			},
			Pairs{
				{"bar", "baz"},
				{"foo", "a b c d e f"},
			},
		},
	}

	for _, tc := range testCases {
		sort.Sort(tc.input)
		require.Equal(t, tc.exp, tc.input)
	}
}

func TestPairsFormat(t *testing.T) {
	testCases := []struct {
		input Pairs
		exp   string
	}{
		{
			Pairs{
				{"foo", "a b c d e f"},
				{"bar", "baz"},
			},
			`bar=baz foo="a b c d e f"`,
		},
		{
			Pairs{
				{"a", `"bar"baz"`},
			},
			`a="\"bar\"baz\""`,
		},
	}

	for _, tc := range testCases {
		sort.Sort(tc.input)
		s := tc.input.Format()
		require.Equal(t, tc.exp, s)
	}
}
