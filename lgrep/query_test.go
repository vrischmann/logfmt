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

func TestQueryMatchKey(t *testing.T) {
	testCases := []struct {
		input []string
		qry   Query
		exp   bool
	}{
		{
			[]string{"a", "b", "c"},
			Query{key: "foo"},
			false,
		},
		{
			[]string{"foo"},
			Query{key: "foo"},
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
				{key: "foo", value: "bar"},
			},
			true,
		},
		{
			"foo=bar bar=baz",
			nil,
			Queries{
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
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
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
			},
			true,
		},
		{
			"foo=bar",
			&QueryOption{
				Or: false,
			},
			Queries{
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
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
				{key: "foo", value: "bar"},
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
				{key: "foo", value: "bar"},
			},
			false,
		},
		{
			"foo=bar bar=baz",
			&QueryOption{
				Reverse: true,
			},
			Queries{
				{key: "foo", value: "bar"},
			},
			false,
		},
		{
			"foo=bar bar=baz",
			&QueryOption{
				Reverse: true,
			},
			Queries{
				{key: "foo", value: "bar"},
				{key: "a", value: "b"},
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
				{key: "b", value: "c"},
				{key: "a", value: "b"},
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
		q     Queries
		exp   bool
	}{
		{
			[]string{"a", "b", "c"},
			Queries{
				{key: "foo"},
			},
			false,
		},
		{
			[]string{"foo"},
			Queries{
				{key: "foo"},
			},
			true,
		},
		{
			[]string{"foo", "bar"},
			Queries{
				{key: "foo"},
				{key: "ba"},
			},
			false,
		},
		{
			[]string{"foo", "bar", "abcd"},
			Queries{
				{key: "foo"},
				{key: "bar"},
			},
			true,
		},
	}

	for _, tc := range testCases {
		res := tc.q.MatchKeys(tc.input)
		require.Equal(t, tc.exp, res)
	}
}
