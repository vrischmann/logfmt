package internal

import (
	"compress/gzip"
	"io"
	"log"
	"os"
)

func isGzip(r io.Reader) (bool, error) {
	var buf [3]byte

	_, err := io.ReadFull(r, buf[:])
	switch {
	case err == io.EOF:
		return false, nil
	case err != nil:
		return false, err
	}

	const (
		gzipID1     = 0x1f
		gzipID2     = 0x8b
		gzipDeflate = 8
	)

	gzip := buf[0] == gzipID1 && buf[1] == gzipID2 && buf[2] == gzipDeflate

	return gzip, nil
}

// getReader returns a reader for the file, handling gzip compression if necessary
func getReader(filename string) (io.Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	gz, err := isGzip(f)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	if gz {
		return gzip.NewReader(f)
	}

	return f, nil
}

// GetInput returns a single Reader concatenating all appropriate sources.
// If multiples files are provided, it's a reader concatenating every files in the order provided.
// If no files is provided, the input is the standard input.
//
// Additionally, this functions takes care of ungzipping files if necessary.
func GetInput(args []string) io.Reader {
	var input io.Reader = os.Stdin

	if len(args) > 0 {
		readers := make([]io.Reader, 0, len(args))
		for _, filename := range args {
			rd, err := getReader(filename)
			if err != nil {
				log.Fatal(err)
			}

			readers = append(readers, rd)
		}

		input = io.MultiReader(readers...)
	}

	return input
}
