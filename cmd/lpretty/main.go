package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
)

const transformOperator = "::"

func extractTransforms(args []string) transforms {
	var res transforms

	if len(args) == 0 {
		return transforms{&stripKeyTransform{}}
	}

	if len(args) == 1 && strings.EqualFold(args[0], "json") {
		return transforms{mergeToJSONTransform{}}
	}

	for _, arg := range args {
		if strings.Contains(arg, transformOperator) {
			tokens := strings.Split(arg, transformOperator)
			res = append(res, &singlePairTransform{
				key: tokens[0],
				typ: tokens[1],
			})

			continue
		}

		res = append(res, &stripKeyTransform{key: arg})

		break
	}

	return res
}

func main() {
	flag.Parse()

	args := flag.Args()
	transforms := extractTransforms(args)
	if len(args) > 0 {
		args = args[len(transforms):]
	}

	//

	inputs := internal.GetInputs(args)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		for scanner.Scan() {
			line := scanner.Text()
			pairs := logfmt.Split(line)

			//

			result := transforms.Apply(pairs)
			switch v := result.(type) {
			case logfmt.Pairs:
				for _, pair := range v {
					buf = append(buf, []byte(pair.Value)...)
				}
				buf = append(buf, '\n')

			case []byte:
				buf = append(buf, v...)
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
