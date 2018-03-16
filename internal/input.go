package internal

import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type Input struct {
	Name   string
	Reader io.Reader
	Data   []byte
}

// GetInputs returns all inputs defined in `args`.
// This function takes care of gunzipping files if necessary, as well as recursively reading
// files from a directory.
func GetInputs(args []string) []Input {
	if len(args) == 0 {
		return []Input{{"stdin", os.Stdin, nil}}
	}

	inputs := make([]Input, 0, len(args))
	for _, source := range args {
		filenames, err := gatherFilenames(source)
		if err != nil {
			log.Fatal(err)
		}

		for _, filename := range filenames {
			input, err := getInput(filename)
			if err != nil {
				log.Fatal(err)
			}

			inputs = append(inputs, *input)
		}
	}

	return inputs
}

func getInput(filename string) (*Input, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	gz, err := isGzip(f)
	if err != nil {
		return nil, err
	}

	switch {
	case gz:
		rd, err := getReader(f, true)
		if err != nil {
			return nil, err
		}

		return &Input{
			Name:   filename,
			Reader: rd,
			Data:   nil,
		}, nil

	default:
		data, err := getMmappedData(f)
		if err != nil {
			return nil, err
		}

		return &Input{
			Name:   filename,
			Reader: nil,
			Data:   data,
		}, nil
	}
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

func getMmappedData(f *os.File) ([]byte, error) {
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_PRIVATE)
}

// getReader returns a reader for the file, handling gzip compression if necessary
func getReader(f *os.File, gz bool) (io.Reader, error) {
	_, err := f.Seek(0, io.SeekStart)
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
