package internal

import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Input struct {
	Name   string
	Reader io.Reader
}

// GetInputs returns all inputs defined in `args`.
// This function takes care of gunzipping files if necessary, as well as recursively reading
// files from a directory.
func GetInputs(args []string) []Input {
	if len(args) == 0 {
		return []Input{{"stdin", os.Stdin}}
	}

	inputs := make([]Input, 0, len(args))
	for _, source := range args {
		filenames, err := gatherFilenames(source)
		if err != nil {
			log.Fatal(err)
		}

		for _, filename := range filenames {
			rd, err := getReader(filename)
			if err != nil {
				log.Fatal(err)
			}

			inputs = append(inputs, Input{
				Name:   filename,
				Reader: rd,
			})
		}
	}

	return inputs
}

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

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	if gz {
		return gzip.NewReader(f)
	}

	return f, nil
}

func gatherFilenames(filename string) ([]string, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return []string{filename}, nil
	}

	files := make([]string, 0, 32)
	err = filepath.Walk(filename, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
