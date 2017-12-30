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

func TestQueryMatch(t *testing.T) {
	testCases := []struct {
		input string
		qry   Query
		exp   bool
	}{
		{
			"foo=bar",
			Query{key: "foo", value: "bar"},
			true,
		},
		{
			"foo=abcdefgh",
			Query{key: "foo", value: "def", fuzzy: true},
			true,
		},
		{
			"foo=012494585",
			Query{key: "foo", regexp: regexp.MustCompile("[0-9]+")},
			true,
		},
		// non matches
		{
			"foo=bar",
			Query{key: "ab", value: "cd"},
			false,
		},
		{
			"foo=ab",
			Query{key: "foo", value: "cd"},
			false,
		},
		{
			"foo=paris",
			Query{key: "foo", value: "rd", fuzzy: true},
			false,
		},
		{
			"foo=012494585",
			Query{key: "foo", regexp: regexp.MustCompile("[a-z]+")},
			false,
		},
	}

	for _, tc := range testCases {
		res := tc.qry.Match(tc.input)
		require.Equal(t, tc.exp, res)
	}
}

func TestQueriesMatch(t *testing.T) {
	testCases := []struct {
		input string
		or    bool
		q     Queries
		exp   bool
	}{
		{
			"foo=bar bar=baz",
			false,
			Queries{
				{key: "foo", value: "bar"},
			},
			true,
		},
		{
			"foo=bar bar=baz",
			false,
			Queries{
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
			},
			true,
		},
		{
			"foo=bar",
			true,
			Queries{
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
			},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.q.Match(tc.or, tc.input)
		require.Equal(t, tc.exp, res)
	}
}
