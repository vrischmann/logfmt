# logfmt

## Library

This package provides functions and types to work with logfmt log data.

Right now you can:

 * parse a logline with [logfmt.Split](https://godoc.org/github.com/vrischmann/logfmt#Split)
 * format key value pairs with [Pairs.Format](https://godoc.org/github.com/vrischmann/logfmt#Pairs.Format) or [Pairs.AppendFormat](https://godoc.org/github.com/vrischmann/logfmt#Pairs.AppendFormat).

## Tools

The goal of logfmt is also to provide a couple of tools to easily parse and integrate log data into a Unix shell pipeline.

No tools are currently developed, but the following is what I will do first.

Note this is not a definitive list, it's likely higher-level tools will be added later.

### lgrep

Works like `grep` but aware of the semantics of logfmt.

For example, you might want to match on a single key, or the value is usually quoted. This tool will deal with that for you.

It would work something like this:

    lgrep foo=bar foo=baz file.log  // implicit OR matches. Matches are strict.
    lgrep foo~bar file.log          // fuzzy matching.
    lgrep foo~=bar file.log         // regex matching.
    lgrep -v foo=bar                // like grep, -v reverses the matching.

### lcut

Remove fields from each log line.

Useful to unclutter logs for example.

It would work something like this:

    lcut foo bar baz file.log      // implicit AND. Removes all 3 fields.
    lcut -v foo file.log           // reverses the matching, meaning only foo will be left in the fields.
