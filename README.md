# logfmt

## Install

    go get -u github.com/vrischmann/logfmt/cmd/lgrep
    go get -u github.com/vrischmann/logfmt/cmd/lcut
    go get -u github.com/vrischmann/logfmt/cmd/lpretty
    go get -u github.com/vrischmann/logfmt/cmd/lsort

## Library

This package provides functions and types to work with logfmt log data.

Right now you can:

 * parse a logline with [logfmt.Split](https://godoc.org/github.com/vrischmann/logfmt#Split)
 * format key value pairs with [Pairs.Format](https://godoc.org/github.com/vrischmann/logfmt#Pairs.Format) or [Pairs.AppendFormat](https://godoc.org/github.com/vrischmann/logfmt#Pairs.AppendFormat).

## Tools

The goal of logfmt is also to provide a couple of tools to easily parse and integrate log data into a Unix shell pipeline.

Note this is not a definitive list, it's likely higher-level tools will be added later.

Each tool is extensively documented, just call it with the `--help` flag.

### lgrep

Works like `grep` but aware of the semantics of logfmt.

For example, you might want to match on a single key, or the value is usually quoted. This tool will deal with that for you.

It would work something like this:

    lgrep foo=bar foo=baz file.log  // implicit OR matches. Matches are strict.
    lgrep foo~bar file.log          // fuzzy matching.
    lgrep foo=~bar file.log         // regex matching.
    lgrep -v foo=bar                // like grep, -v reverses the matching.

### lcut

Remove fields from each log line.

Useful to unclutter logs for example.

It would work something like this:

    lcut foo bar baz file.log      // implicit AND. Removes all 3 fields.
    lcut -v foo file.log           // reverses the matching, meaning only foo will be left in the fields.

### lpretty

Prettify fields from each log line.

Works by applying transformations on fields. Currently three transforms are provided:

Two modes are provided, the normal mode which applies one transformation per field provided on the input:

    lpretty error::exception      // treat the error field as containing a Java exception and display it properly.
    lpretty request::json         // treat the request field as containing valid JSON data and display it properly.
    lpretty name id               // no transform so the default is to remove the keys. This will result in the two values concatenated.

The second mode merges the fields in a single JSON output per log line:

    lpretty -M name id            // returns an object with the name and id fields.
    lpretty -M name req::json     // returns an object with the name and the req field as an object instead of a string value.

### lsort

Sort the input lines based on a logfmt field.

It can sort alphabetically:

    $ cat data.txt
    name=Vincent
    name=John
    name=Sophie
    $ cat data.txt | lsort name
    name=John
    name=Sophie
    name=Vincent

Numerically:

    $ cat data.txt
    age=32 name=Vincent
    age=100 name=Naomi
    age=59 name=John
    age=15 name=Sophie
    $ cat data.txt | lsort name -n
    age=15 name=Sophie
    age=32 name=Vincent
    age=59 name=John
    age=100 name=Naomi

Or by duration:

    $ cat foobar.txt
    elapsed=10m22s foo=bar
    elapsed=3m bar=baz
    elapsed=12s baz=qux
    $ cat foobar.txt | lsort -d id
    elapsed=12s baz=qux
    elapsed=3m bar=baz
    elapsed=10m22s foo=bar
