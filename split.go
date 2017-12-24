package logfmt

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

// Split splits a log line according to the logfmt rules and produces key-value pairs.
//
// It correctly handles double-quoted values.
func Split(line string) Pairs {
	parser := newPairParser()
	parser.rd.Reset(line)

	fn := parser.readKey

	for {
		nextFn := fn()
		if nextFn == nil {
			break
		}

		fn = nextFn
	}

	return parser.pairs
}

type pairParser struct {
	rd  *strings.Reader
	buf *bytes.Buffer

	pairs       Pairs
	currentPair Pair
}

func newPairParser() *pairParser {
	return &pairParser{
		rd:  new(strings.Reader),
		buf: new(bytes.Buffer),
	}
}

type stateFn func() stateFn

func (p *pairParser) maybeMoveBufToValue(unquote bool) {
	if p.buf.Len() > 0 {
		p.currentPair.Value = p.buf.String()
		if unquote {
			p.currentPair.Value, _ = strconv.Unquote(p.currentPair.Value)
		}
		p.pairs = append(p.pairs, p.currentPair)
	}
}

func (p *pairParser) readKey() stateFn {
	p.consumeWhitespace()

	p.currentPair.Key = ""
	p.currentPair.Value = ""
	p.buf.Reset()

	for {
		ch := p.readRune()
		switch ch {
		case eof:
			return nil
		case '=':
			p.currentPair.Key = p.buf.String()
			return p.readValue
		default:
			p.buf.WriteRune(ch)
		}
	}
}

func (p *pairParser) readValue() stateFn {
	p.consumeWhitespace()

	p.buf.Reset()

	for {
		ch := p.readRune()
		switch ch {
		case eof:
			p.maybeMoveBufToValue(false)
			return nil
		case ' ':
			p.maybeMoveBufToValue(false)
			return p.readKey
		case '"':
			return p.readQuotedValue
		default:
			p.buf.WriteRune(ch)
		}
	}
}

// readQuotedValue reads a value that is double-quoted.
// Leverages https://golang.org/pkg/strconv/#Unquote.
func (p *pairParser) readQuotedValue() stateFn {
	p.buf.Reset()

	p.buf.WriteRune('"')

	var ignoreNextQuote bool

loop:
	for {
		ch := p.readRune()
		switch ch {
		case eof:
			p.maybeMoveBufToValue(false)
			return nil
		case '\\':
			p.buf.WriteRune(ch)
			ignoreNextQuote = true
		case '"':
			p.buf.WriteRune(ch)
			if ignoreNextQuote {
				ignoreNextQuote = false
				continue loop
			}

			p.maybeMoveBufToValue(true)

			return p.readKey
		default:
			p.buf.WriteRune(ch)
		}
	}
}

var eof = rune(0)

func (p *pairParser) readRune() rune {
	ch, _, err := p.rd.ReadRune()
	switch {
	case err == io.EOF:
		return eof
	default:
		return ch
	}
}

func (p *pairParser) consumeWhitespace() {
	for {
		ch := p.readRune()
		if ch != ' ' {
			p.rd.UnreadRune()
			return
		}
	}
}
