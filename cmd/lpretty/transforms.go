package main

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/vrischmann/logfmt"
)

type transform interface {
	Apply(logfmt.Pairs) interface{}
}

type stripKeyTransform struct {
	key string
}

func (t *stripKeyTransform) Apply(pairs logfmt.Pairs) interface{} {
	if len(pairs) <= 0 {
		return pairs
	}

	if pairs[0].Key != t.key {
		return pairs
	}

	pairs[0].Key = ""

	return pairs[:1]
}

type singlePairTransform struct {
	key string
	typ string
}

func (t *singlePairTransform) Apply(pairs logfmt.Pairs) interface{} {
	ret := make(logfmt.Pairs, len(pairs))

loop:
	for i, pair := range pairs {
		if pair.Key != t.key {
			continue
		}

		switch t.typ {
		case "json":
			buf := new(bytes.Buffer)
			err := json.Indent(buf, []byte(pair.Value), "", "  ")
			if err != nil {
				continue loop
			}

			pair.Value = buf.String()

		case "exception":
			data, err := strconv.Unquote(pair.Value)
			if err != nil {
				continue loop
			}

			pair.Value = data
		}

		ret[i] = pair
	}

	return ret
}

type mergeToJSONTransform struct {
}

func (t mergeToJSONTransform) Apply(pairs logfmt.Pairs) interface{} {
	obj := make(map[string]interface{})
	for _, pair := range pairs {
		obj[pair.Key] = pair.Value
	}

	data, _ := json.Marshal(obj)

	return data
}

type transforms []transform

func (t *transforms) Apply(pairs logfmt.Pairs) interface{} {
	ret := make(logfmt.Pairs, len(pairs))
	copy(ret, pairs)

	for _, tr := range *t {
		tmp := tr.Apply(ret)
		v, ok := tmp.(logfmt.Pairs)
		if !ok {
			panic("can only chain transforms which produce pairs, not a merged result")
		}

		ret = v
	}

	return ret
}
