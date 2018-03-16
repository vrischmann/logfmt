package main

import (
	"bufio"
	"bytes"
	"io"
	"reflect"
	"unsafe"

	"github.com/vrischmann/logfmt/internal/flags"
)

type scanner interface {
	Scan() bool
	String() string
	Err() error
}

type noCopyScanner struct {
	data      []byte
	strHeader *reflect.StringHeader

	cur  []byte
	next []byte
}

func newNoCopyScanner(data []byte) scanner {
	return &noCopyScanner{
		data:      data,
		strHeader: new(reflect.StringHeader),
		cur:       data,
	}
}

func (s *noCopyScanner) Scan() bool {
	idx := bytes.IndexRune(s.cur, '\n')
	if idx == -1 {
		return false
	}

	s.next = s.cur[:idx]
	s.cur = s.cur[idx+1:]

	return true
}

func (s *noCopyScanner) String() string {
	data := s.next

	s.strHeader.Data = uintptr(unsafe.Pointer(&data[0]))
	s.strHeader.Len = len(data)

	return *(*string)(unsafe.Pointer(s.strHeader))
}

func (s *noCopyScanner) Err() error {
	return nil
}

type stdScanner struct {
	s         *bufio.Scanner
	strHeader *reflect.StringHeader
}

func newStdScanner(rd io.Reader) scanner {
	s := bufio.NewScanner(rd)
	s.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))

	return &stdScanner{
		s:         s,
		strHeader: new(reflect.StringHeader),
	}
}

func (s *stdScanner) Scan() bool {
	return s.s.Scan()
}

func (s *stdScanner) String() string {
	data := s.s.Bytes()

	s.strHeader.Data = uintptr(unsafe.Pointer(&data[0]))
	s.strHeader.Len = len(data)

	return *(*string)(unsafe.Pointer(s.strHeader))
}

func (s *stdScanner) Err() error {
	return s.s.Err()
}
