package logfmt

import (
	"bufio"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
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

func TestSplitInto(t *testing.T) {

	testCases := []string{
		"city=Lyon name=Vincent age=123",
		"city=Lyon name=Vincent age=123",
		"city=Lyon name=Vincent age=123",
	}

	var (
		parser PairParser
		pairs  = make(Pairs, 32)
	)
	for _, tc := range testCases {
		pairs = parser.SplitInto(tc, pairs)
		require.Equal(t, 3, len(pairs))
	}
}

func BenchmarkSplit(b *testing.B) {
	const line = `city=Lyon name=Vincent age=123 latitude=0.2982902490 longitude=95.2023904 str="foo bar baz" json="{\"Foo\":\"foo\",\"Bar\":\"bar\",\"Baz\":{\"A\":12,\"B\":4540,\"C\":{\"Opened\":true}}}"`

	var parser PairParser
	for i := 0; i < b.N; i++ {
		pairs := parser.Split(line)
		if len(pairs) <= 0 {
			b.Fatal("should have at least one pair")
		}
	}
}

// If you want to run this, generate a file with this command
//    go run gen_test_log.go -size 10 -output /tmp/logfmt_benchmark_data.log
func BenchmarkSplitFile(b *testing.B) {
	const dataFile = "/tmp/logfmt_benchmark_data.log"

	f, err := os.Open(dataFile)
	if err != nil {
		logrus.WithError(err).WithField("data-file", dataFile).Warn("unable to run benchmark because data file is not present")
		b.Skip()
	}

	b.ResetTimer()

	var parser PairParser
	for i := 0; i < b.N; i++ {
		_, err := f.Seek(0, os.SEEK_SET)
		require.NoError(b, err)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			pairs := parser.Split(line)
			if len(pairs) <= 0 {
				b.Fatal("should have at least one pair")
			}
		}
	}
}
