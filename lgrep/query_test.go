package lgrep

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkQueryFuzzyPresent(b *testing.B) {
	q := newQuery("house")
	q.fuzzy = true
	line := strings.Repeat("house=foobar ", 1000)
	line = strings.Repeat("foo=bar ", 1000) + line

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Match(line)
	}
}

func BenchmarkQueryFuzzyNotPresent(b *testing.B) {
	q := newQuery("house")
	q.fuzzy = true
	line := strings.Repeat("foo=bar ", 1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = q.Match(line)
	}
}

func BenchmarkQueryNotFuzzyPresent(b *testing.B) {
	q := newQuery("house")
	line := strings.Repeat("house=foobar ", 1000)
	line = strings.Repeat("foo=bar ", 1000) + line

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Match(line)
	}
}

func BenchmarkQueryNotFuzzyNotPresent(b *testing.B) {
	q := newQuery("house")
	line := strings.Repeat("foo=bar ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Match(line)
	}
}

func mkq(key, value string) Query {
	q := newQuery(key)
	q.value = value
	return q
}

func mkfq(key, value string) Query {
	q := mkq(key, value)
	q.fuzzy = true
	return q
}

func mkrq(key, value, re string) Query {
	q := mkq(key, value)
	q.regexp = regexp.MustCompile(re)
	return q
}

func TestQueryMatch(t *testing.T) {
	testCases := []struct {
		input string
		qry   Query
		exp   bool
	}{
		{
			"foo=bar",
			mkq("foo", "bar"),
			true,
		},
		{
			"foo=abcdefgh",
			mkfq("foo", "def"),
			true,
		},
		{
			"foo=012494585",
			mkrq("foo", "", "[0-9]+"),
			true,
		},
		// non matches
		{
			"foo=bar",
			mkq("ab", "cd"),
			false,
		},
		{
			"foo=ab",
			mkq("foo", "cd"),
			false,
		},
		{
			"foo=paris",
			mkfq("foo", "rd"),
			false,
		},
		{
			"foo=012494585",
			mkrq("foo", "", "[a-z]+"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			res := tc.qry.Match(tc.input)
			require.Equal(t, tc.exp, res)
		})
	}
}

func TestQueryMatchKey(t *testing.T) {
	testCases := []struct {
		input []string
		qry   Query
		exp   bool
	}{
		{
			[]string{"a", "b", "c"},
			mkq("foo", ""),
			false,
		},
		{
			[]string{"foo"},
			mkq("foo", ""),
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.qry.MatchKeys(tc.input)
		require.Equal(t, tc.exp, res)
	}
}

func TestQueriesMatch(t *testing.T) {
	testCases := []struct {
		input string
		opt   *QueryOption
		q     Queries
		exp   bool
	}{
		{
			"foo=bar bar=baz",
			nil,
			Queries{
				mkq("foo", "bar"),
			},
			true,
		},
		{
			"foo=bar bar=baz",
			nil,
			Queries{
				mkq("foo", "bar"),
				mkq("bar", "baz"),
			},
			true,
		},
		//
		{
			"foo=bar",
			&QueryOption{
				Or: true,
			},
			Queries{
				mkq("foo", "bar"),
				mkq("bar", "baz"),
			},
			true,
		},
		{
			"foo=bar",
			&QueryOption{
				Or: false,
			},
			Queries{
				mkq("foo", "bar"),
				mkq("bar", "baz"),
			},
			false,
		},
		//
		{
			"foo=bar",
			&QueryOption{
				Reverse: true,
			},
			Queries{
				mkq("foo", "bar"),
			},
			false,
		},
		{
			"foo=bar",
			&QueryOption{
				Or:      true,
				Reverse: true,
			},
			Queries{
				mkq("foo", "bar"),
			},
			false,
		},
		{
			"foo=bar bar=baz",
			&QueryOption{
				Reverse: true,
			},
			Queries{
				mkq("foo", "bar"),
			},
			false,
		},
		{
			"foo=bar bar=baz",
			&QueryOption{
				Reverse: true,
			},
			Queries{
				mkq("foo", "bar"),
				mkq("a", "b"),
			},
			true,
		},
		//
		{
			"foo=bar bar=baz",
			&QueryOption{
				Or: true,
			},
			Queries{
				mkq("b", "c"),
				mkq("a", "b"),
			},
			false,
		},
	}

	for _, tc := range testCases {
		res := tc.q.Match(tc.input, tc.opt)
		require.Equal(t, tc.exp, res)
	}
}

func TestQueriesMatchKeys(t *testing.T) {
	testCases := []struct {
		input []string
		opt   *QueryOption
		q     Queries
		exp   bool
	}{
		{
			[]string{"a", "b", "c"},
			nil,
			Queries{
				mkq("foo", ""),
			},
			false,
		},
		{
			[]string{"foo"},
			nil,
			Queries{mkq("foo", "")},
			true,
		},
		{
			[]string{"foo", "bar"},
			nil,
			Queries{
				mkq("foo", ""),
				mkq("ba", ""),
			},
			false,
		},
		{
			[]string{"foo", "bar", "abcd"},
			nil,
			Queries{
				mkq("foo", ""),
				mkq("bar", ""),
			},
			true,
		},
		//
		{
			[]string{"a", "b"},
			&QueryOption{
				Reverse: true,
			},
			Queries{
				mkq("foo", ""),
			},
			true,
		},
		{
			[]string{"a"},
			&QueryOption{
				Reverse: true,
			},
			Queries{
				mkq("a", ""),
			},
			false,
		},
	}

	for _, tc := range testCases {
		res := tc.q.MatchKeys(tc.input, tc.opt)
		require.Equal(t, tc.exp, res)
	}
}
