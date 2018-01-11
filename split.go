package logfmt

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

// Split splits a log line according to the logfmt rules and produces key-value pairs.
// It is a convenience function which does the same as PairParser.Split(line).
func Split(line string) Pairs {
	var parser PairParser
	return parser.Split(line)
}

// PairParser is a parser of key-value pairs. It parses a logline according to the logfmt rules.
type PairParser struct {
	rd   *strings.Reader
	buf  *bytes.Buffer
	done bool

	pairs       Pairs
	currentPair Pair
}

// Split splits a log line according to the logfmt rules and produces key-value pairs.
func (p *PairParser) Split(line string) Pairs {
	var pairs Pairs
	return p.SplitInto(line, pairs)
}

// SplitInto splits a log line according to the logfmt rules and produces key-value pairs.
// This function appends the pairs to `pairs` and return the slice truncated.
func (p *PairParser) SplitInto(line string, pairs Pairs) Pairs {
	if p.rd == nil {
		p.rd = new(strings.Reader)
		p.buf = new(bytes.Buffer)
	}
	p.done = false
	p.rd.Reset(line)
	p.pairs = pairs[:0]

	for !p.done {
		p.readKey()
		p.readValue()
	}

	return p.pairs
}

func (p *PairParser) maybeMoveBufToValue(unquote bool) {
	if p.buf.Len() > 0 {
		p.currentPair.Value = p.buf.String()
		if unquote {
			p.currentPair.Value, _ = strconv.Unquote(p.currentPair.Value)
		}
		p.pairs = append(p.pairs, p.currentPair)
	}
}

func (p *PairParser) readKey() {
	p.consumeWhitespace()

	p.currentPair.Key = ""
	p.currentPair.Value = ""
	p.buf.Reset()

	for {
		ch := p.readRune()
		switch ch {
		case eof:
			p.done = true
			return
		case '=':
			p.currentPair.Key = p.buf.String()
			return
		default:
			p.buf.WriteRune(ch)
		}
	}
}

func (p *PairParser) readValue() {
	p.consumeWhitespace()

	p.buf.Reset()

	for {
		ch := p.readRune()
		switch ch {
		case eof:
			p.maybeMoveBufToValue(false)
			p.done = true
			return
		case ' ':
			p.maybeMoveBufToValue(false)
			return
		case '"':
			p.readQuotedValue()
			return
		default:
			p.buf.WriteRune(ch)
		}
	}
}

// readQuotedValue reads a value that is double-quoted.
// Leverages https://golang.org/pkg/strconv/#Unquote.
func (p *PairParser) readQuotedValue() {
	p.buf.Reset()

	p.buf.WriteRune('"')

	for {
		ch := p.readRune()
		switch {
		case ch == eof:
			p.maybeMoveBufToValue(false)
			p.done = true
			return
		case ch == '\\':
			nextCh := p.readRune()
			p.buf.WriteRune(ch)
			p.buf.WriteRune(nextCh)
		case ch == '"':
			p.buf.WriteRune(ch)
			p.maybeMoveBufToValue(true)
			return
		default:
			p.buf.WriteRune(ch)
		}
	}
}

var eof = rune(0)

func (p *PairParser) readRune() rune {
	ch, _, err := p.rd.ReadRune()
	switch {
	case err == io.EOF:
		return eof
	default:
		return ch
	}
}

func (p *PairParser) consumeWhitespace() {
	for {
		ch := p.readRune()
		if ch != ' ' {
			p.rd.UnreadRune()
			return
		}
	}
}
