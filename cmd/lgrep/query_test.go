package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkQuery(b *testing.B) {
	q := query{
		key:   "house",
		fuzzy: true,
	}

	b.Run("present", func(b *testing.B) {
		b.ResetTimer()
		line := strings.Repeat("house=foobar ", 1000)
		line = strings.Repeat("foo=bar ", 1000) + line
		for i := 0; i < b.N; i++ {
			_ = q.Match(line)
		}
	})

	b.Run("not-present", func(b *testing.B) {
		b.ResetTimer()
		line := strings.Repeat("foo=bar ", 1000)
		for i := 0; i < b.N; i++ {
			_ = q.Match(line)
		}
	})
}

func TestQueryMatch(t *testing.T) {
	testCases := []struct {
		input string
		qry   query
		exp   bool
	}{
		{
			"foo=bar",
			query{key: "foo", value: "bar"},
			true,
		},
		{
			"foo=abcdefgh",
			query{key: "foo", value: "def", fuzzy: true},
			true,
		},
		{
			"foo=012494585",
			query{key: "foo", regexp: regexp.MustCompile("[0-9]+")},
			true,
		},
		// non matches
		{
			"foo=bar",
			query{key: "ab", value: "cd"},
			false,
		},
		{
			"foo=ab",
			query{key: "foo", value: "cd"},
			false,
		},
		{
			"foo=paris",
			query{key: "foo", value: "rd", fuzzy: true},
			false,
		},
		{
			"foo=012494585",
			query{key: "foo", regexp: regexp.MustCompile("[a-z]+")},
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
		q     queries
		exp   bool
	}{
		{
			"foo=bar bar=baz",
			false,
			queries{
				{key: "foo", value: "bar"},
			},
			true,
		},
		{
			"foo=bar bar=baz",
			false,
			queries{
				{key: "foo", value: "bar"},
				{key: "bar", value: "baz"},
			},
			true,
		},
		{
			"foo=bar",
			true,
			queries{
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
