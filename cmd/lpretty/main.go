package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
)

const transformOperator = "::"

func extractTransform() (transform, []string) {
	args := flag.Args()

	if *flMerge {
		return newMergeToJSONTransform(args), nil
	}

	//

	var res transforms

	if len(args) == 0 {
		return &stripKeyTransform{}, args
	}

	for _, arg := range args {
		t := newSinglePairTransform(arg)
		switch {
		case t != nil:
			res = append(res, t)
		default:
			res = append(res, &stripKeyTransform{key: arg})
		}

		args = args[1:]
	}

	return res, args
}

var (
	flMerge = flag.Bool("merge", false, "Merge the fields to a JSON object")
)

func main() {
	flag.Parse()

	transform, args := extractTransform()

	//

	inputs := internal.GetInputs(args)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			//

			result := transform.Apply(pairs)
			if result == nil {
				continue
			}

			switch v := result.(type) {
			case logfmt.Pairs:
				for _, pair := range v {
					buf = append(buf, []byte(pair.Value)...)
				}
				buf = append(buf, '\n')

			case []byte:
				buf = append(buf, v...)
				buf = append(buf, '\n')

			default:
				panic(fmt.Errorf("invalid result type %T", result))
			}

			//

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
