package logfmt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplit(t *testing.T) {
	testCases := []struct {
		input string
		exp   Pairs
	}{
		{
			"ab=cd",
			Pairs{
				{"ab", "cd"},
			},
		},
		{
			"foo=bar 1=2    a=b",
			Pairs{
				{"foo", "bar"},
				{"1", "2"},
				{"a", "b"},
			},
		},
		{
			`str="foo bar baz" json="{\"Foo\":\"foo\",\"Bar\":\"bar\",\"Baz\":{\"A\":12,\"B\":4540,\"C\":{\"Opened\":true}}}"`,
			Pairs{
				{"str", "foo bar baz"},
				{"json", `{"Foo":"foo","Bar":"bar","Baz":{"A":12,"B":4540,"C":{"Opened":true}}}`},
			},
		},
		{
			`json="\"{\\\"Foo\\\":\\\"foo\\\",\\\"Bar\\\":\\\"bar\\\",\\\"Baz\\\":{\\\"A\\\":12,\\\"B\\\":4540,\\\"C\\\":{\\\"Opened\\\":true}}}\""`,
			Pairs{
				{"json", `"{\"Foo\":\"foo\",\"Bar\":\"bar\",\"Baz\":{\"A\":12,\"B\":4540,\"C\":{\"Opened\":true}}}"`},
			},
		},
	}

	for _, tc := range testCases {
		res := Split(tc.input)
		require.Equal(t, tc.exp, res)
	}
}
