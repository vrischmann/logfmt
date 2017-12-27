package internal

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsGzip(t *testing.T) {
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)

	io.WriteString(w, "foobar")
	require.NoError(t, w.Flush())

	ok, err := isGzip(buf)
	require.NoError(t, err)
	require.True(t, ok)

	buf.Reset()

	ok, err = isGzip(buf)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestGetReader(t *testing.T) {
	testCases := []struct {
		data   string
		mkgzip bool
	}{
		{"foobar", false},
		{"foobar", true},
	}

	for _, tc := range testCases {
		filename := mkFile(t, "", "logfmt", tc.data, tc.mkgzip)

		rd, err := getReader(filename)
		require.NoError(t, err)

		if tc.mkgzip {
			_, ok := rd.(*gzip.Reader)
			require.True(t, ok)
		} else {
			_, ok := rd.(*os.File)
			require.True(t, ok)
		}

		data, err := ioutil.ReadAll(rd)
		require.NoError(t, err)
		require.Equal(t, tc.data, string(data))
	}
}

func mkFile(t *testing.T, dir, prefix string, data string, mkgzip bool) string {
	f, err := ioutil.TempFile(dir, prefix)
	require.NoError(t, err)
	defer f.Close()

	if mkgzip {
		w := gzip.NewWriter(f)
		_, err := io.WriteString(w, data)
		require.NoError(t, err)
		err = w.Close()
		require.NoError(t, err)

		return f.Name()
	}

	io.WriteString(f, data)

	return f.Name()
}

func TestGetInput(t *testing.T) {
	filenames := []string{
		mkFile(t, "", "logfmt", "foobar1", false),
		mkFile(t, "", "logfmt", "foobar2", true),
		mkFile(t, "", "logfmt", "foobar3", false),
		mkFile(t, "", "logfmt", "foobar4", true),
	}

	input := GetInput(filenames)

	const exp = "foobar1foobar2foobar3foobar4"

	data, err := ioutil.ReadAll(input)
	require.NoError(t, err)
	require.Equal(t, exp, string(data))
}

func TestGetInputDirectory(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "logfmt_test")
	require.NoError(t, os.MkdirAll(dir, 0755))
	defer func() {
		os.RemoveAll(dir)
	}()

	mkFile(t, dir, "logfmt1", "foobar1", false)
	mkFile(t, dir, "logfmt2", "foobar2", true)
	mkFile(t, dir, "logfmt3", "foobar3", false)
	mkFile(t, dir, "logfmt4", "foobar4", true)

	input := GetInput([]string{dir})

	const exp = "foobar1foobar2foobar3foobar4"

	data, err := ioutil.ReadAll(input)
	require.NoError(t, err)
	require.Equal(t, exp, string(data))
}
