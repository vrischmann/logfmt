package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
)

type transformType int

const (
	jsonTransform          transformType = 1
	javaExceptionTransform transformType = 2
)

func transformTypeFromString(s string) transformType {
	switch {
	case strings.EqualFold(s, "json"):
		return jsonTransform
	case strings.EqualFold(s, "exception"):
		return javaExceptionTransform
	default:
		return 0
	}
}

type transform struct {
	key string
	typ transformType
}

func (t transform) Apply(pair logfmt.Pair) logfmt.Pair {
	switch t.typ {
	case jsonTransform:
		buf := new(bytes.Buffer)
		err := json.Indent(buf, []byte(pair.Value), "", "  ")
		if err != nil {
			return pair
		}

		pair.Value = buf.String()

	case javaExceptionTransform:
		data, err := strconv.Unquote(pair.Value)
		if err != nil {
			return pair
		}

		pair.Value = data
	}

	return pair
}

type transforms []transform

func (t transforms) Apply(pairs logfmt.Pairs) logfmt.Pairs {
	res := make(logfmt.Pairs, 0, len(pairs))
	for _, transform := range t {
		for _, pair := range pairs {
			if pair.Key == transform.key {
				pair = transform.Apply(pair)
				res = append(res, pair)
			}
		}
	}
	return res
}

const transformOperator = "::"

func extractTransforms(args []string) transforms {
	var res transforms

	for _, arg := range args {
		if strings.Contains(arg, transformOperator) {
			tokens := strings.Split(arg, transformOperator)
			res = append(res, transform{
				key: tokens[0],
				typ: transformTypeFromString(tokens[1]),
			})

			continue
		}

		break
	}

	return res
}

func main() {
	flag.Parse()

	args := flag.Args()
	transforms := extractTransforms(args)
	args = args[len(transforms):]

	//

	inputs := internal.GetInputs(args)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			pairs = transforms.Apply(pairs)

			if len(pairs) <= 0 {
				continue
			}

			for _, pair := range pairs {
				buf = append(buf, []byte(pair.Value)...)
			}
			buf = append(buf, '\n')

			_, err := os.Stdout.Write(buf)
			if err != nil {
				log.Fatal(err)
			}

			buf = buf[:0]
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
}
