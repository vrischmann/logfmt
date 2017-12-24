package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vrischmann/logfmt"
)

func TestCutFields(t *testing.T) {
	testCases := []struct {
		input   string
		cut     []string
		reverse bool
		exp     string
	}{
		{
			"city=Lyon name=Vincent foo=bar",
			[]string{"foo"},
			false,
			"city=Lyon name=Vincent",
		},
		{
			"a=b c=d e=f",
			nil,
			false,
			"a=b c=d e=f",
		},
		{
			"a=b c=d e=f",
			[]string{"a"},
			true,
			"a=b",
		},
	}

	for _, tc := range testCases {
		pairs := logfmt.Split(tc.input)
		f := cutFields(tc.cut)
		pairs = f.CutFrom(tc.reverse, pairs)
		require.Equal(t, tc.exp, pairs.Format())
	}
}
