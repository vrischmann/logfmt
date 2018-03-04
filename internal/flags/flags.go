package flags

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	CPUProfile  string
	MemProfile  string
	MaxLineSize Size
)

type Size int64

const (
	kib = Size(1024)
	mib = kib * 1024
)

func init() {
	MaxLineSize = mib
}

func (z *Size) Set(s string) error {
	switch {
	case strings.HasSuffix(s, "Mib"):
		tmp, err := strconv.ParseInt(s[:len(s)-3], 10, 64)
		if err != nil {
			return err
		}

		*z = Size(tmp) * mib

	case strings.HasSuffix(s, "Kib"):
		tmp, err := strconv.ParseInt(s[:len(s)-3], 10, 64)
		if err != nil {
			return err
		}

		*z = Size(tmp) * kib

	default:
		tmp, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		*z = Size(tmp)
	}

	return nil
}

func (z *Size) String() string {
	switch {
	case *z >= mib:
		return fmt.Sprintf("%fMib", float64(*z)/float64(mib))
	case *z >= kib:
		return fmt.Sprintf("%fKib", float64(*z)/float64(kib))
	default:
		return fmt.Sprintf("%db", *z)
	}
}

func (z Size) Type() string { return "int64" }
