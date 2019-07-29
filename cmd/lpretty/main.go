package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/vrischmann/logfmt"
	"github.com/vrischmann/logfmt/internal"
	"github.com/vrischmann/logfmt/internal/flags"
)

const transformOperator = "::"

func extractTransform(args []string) (transform, []string) {
	if flMerge {
		return newMergeToJSONTransform(flAll, args), nil
	}
	if flNewline {
		return &dummyTransform{}, nil
	}
	if flStripKey {
		return &stripKeyTransform{}, args
	}

	//

	var res transforms

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
func runMain(cmd *cobra.Command, args []string) error {
	transform, args := extractTransform(args)

	//

	inputs := internal.GetInputs(args)

	buf := make([]byte, 0, 4096)
	for _, input := range inputs {
		scanner := bufio.NewScanner(input.Reader)
		scanner.Buffer(make([]byte, int(flags.MaxLineSize)/2), int(flags.MaxLineSize))
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
				if len(v) == 0 {
					break
				}

				switch {
				case flNewline:
					for _, pair := range v {
						buf = append(buf, pair.Key...)
						buf = append(buf, '=')
						buf = append(buf, pair.Value...)
						buf = append(buf, '\n')
					}

					buf = append(buf, '\n')

				default:

					for _, pair := range v {
						buf = append(buf, []byte(pair.Value)...)
					}
					buf = append(buf, '\n')
				}

			case []byte:
				buf = append(buf, v...)
				buf = append(buf, '\n')

			default:
				panic(fmt.Errorf("invalid result type %T", result))
			}

			//

			_, err := os.Stdout.Write(buf)
			if err != nil {
				return err
			}

			buf = buf[:0]
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	rootCmd.Execute()
}

var (
	rootCmd = &cobra.Command{
		Use:   "lpretty [field[::[transform]]",
		Short: "prettify logfmt fields",
		Long: `prettify logfmt fields by applying transformation on them.

Every transform needs to be provided like this:
  <field name>::<transform name>

You can also just provide the field name and in this case the key from the field will be stripped from the output.

Transformations available:
  - json        for fields which contain valid JSON data
  - ulid        for fields which contain a valid ULID
  - error       for fields which contain valid Java exceptions

Examples:

    $ echo 'request="{\"id\":10,\"name\":\"Vincent\"}"' > /tmp/logfmt
    $ cat /tmp/logfmt | lpretty request::json
    {
        "request": {
            "id": 10,
            "name": "Vincent"
        }
    }

There is a second mode where the fields are merged using the flag --merge/-M. In this mode every field listed in the arguments
will be added to a single JSON object. If you add the --all flag it will include _every_ field in the output, applying the
transformations given in the list of arguments (if any).

You can also indicate that a field contains valid JSON data in this mode so that the JSON is embedded directly in the object and
not reencoded as a string value.

Examples:

* The most basic use case: output some fields

	$ echo 'id=10 name=vincent surname=Rischmann age=55' > /tmp/logfmt
	$ cat /tmp/logfmt | lpretty -M id age
	{
		"id": "10",
		"age": "55"
	}

* More complex use case: output some fields but treat one as JSON data

	$ echo 'id=10 name=vincent data="{\"limit\":5000,\"offset\":30,\"table\":\"almanac\"}"' > /tmp/logfmt
	$ cat /tmp/logfmt | lpretty -M id data::json
	{
		"id": "10",
		"data: {
			"limit": 5000,
			"offset": 30,
			"table": "almanac"
		}
	}

* Another relatively basic use case: output all fields in a single JSON object

	$ echo 'id=10 name=vincent surname=Rischmann age=55' > /tmp/logfmt
	$ cat /tmp/logfmt | lpretty -M --all
	{
		"age": "55",
		"id": "10",
		"name": "vincent",
		"surname": "Rischmann"
	}

* If we try that with a field containing JSON data

	$ echo 'id=10 name=vincent data="{\"limit\":5000,\"offset\":30,\"table\":\"almanac\"}"' > /tmp/logfmt
	$ cat /tmp/logfmt | lpretty -M --all
	{
		"data": "{\"limit\":5000,\"offset\":30,\"table\":\"almanac\"}",
		"id": "10",
		"name": "vincent"
	}

* Not the easiest to work, so we can combine both flags

	$ echo 'id=10 name=vincent data="{\"limit\":5000,\"offset\":30,\"table\":\"almanac\"}"' > /tmp/logfmt
	$ cat /tmp/logfmt | lpretty -M --all data::json
	{
		"data": {
			"limit": 5000,
			"offset": 30,
			"table": "almanac"
		},
		"id": "10",
		"name": "vincent"
	}

Finally there's a third mode which strips the key of the first pair and only prints its value.
This is useful when you pipe lpretty to the output of lcut.

Examples:

	$ echo 'id=10 name=vincent surname=Rischmann age=55' > /tmp/logfmt
	$ cat /tmp/logfmt | lcut -v surname | lpretty -S
	Rischmann`,
		RunE: runMain,
	}

	flMerge    bool
	flNewline  bool
	flStripKey bool
	flAll      bool
)

func init() {
	fs := rootCmd.Flags()

	fs.Var(&flags.MaxLineSize, "max-line-size", "Max size in bytes of a line")
	fs.BoolVarP(&flMerge, "merge", "M", false, "Merge all fields in a single JSON object")
	fs.BoolVarP(&flNewline, "newline", "N", false, "Print all fields into its own line")
	fs.BoolVarP(&flStripKey, "strip-key", "S", false, "Strip the key of the first pair and only print the value")
	fs.BoolVar(&flAll, "all", false, "When merging in a single JSON object include all fields, not just the one described in the arguments")
}
