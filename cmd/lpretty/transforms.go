package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"github.com/vrischmann/logfmt"
)

type transform interface {
	Apply(logfmt.Pairs) interface{}
}

type stripKeyTransform struct {
	key string
}

func (t *stripKeyTransform) Apply(pairs logfmt.Pairs) interface{} {
	switch {
	case len(pairs) <= 0:
		return pairs
	case t.key == "":
		pairs[0].Key = ""
		return pairs[0:1]
	}

	for i, pair := range pairs {
		if pair.Key == t.key {
			pair.Key = ""
			return pairs[i : i+1]
		}
	}

	return nil
}

type singlePairTransform struct {
	key string
	typ string
}

func newSinglePairTransform(arg string) *singlePairTransform {
	if !strings.Contains(arg, transformOperator) {
		return nil
	}

	tokens := strings.Split(arg, transformOperator)
	return &singlePairTransform{
		key: tokens[0],
		typ: tokens[1],
	}
}

func (t *singlePairTransform) Apply(pairs logfmt.Pairs) interface{} {
	ret := make(logfmt.Pairs, 0, len(pairs))

loop:
	for _, pair := range pairs {
		if pair.Key != t.key {
			continue
		}

		switch t.typ {
		case "json":
			buf := new(bytes.Buffer)
			err := json.Indent(buf, []byte(pair.Value), "", "  ")
			if err != nil {
				continue loop
			} else {
				pair.Value = buf.String()
			}
		case "ulid":
			id := ulid.MustParse(pair.Value)
			pair.Value = fmt.Sprintf("{original: %s; time: %s}", pair.Value, ulid.Time(id.Time()).UTC().Format(time.RFC3339))
		}

		ret = append(ret, pair)
	}

	return ret
}

type mergeToJSONTransform struct {
	all  bool
	keys map[string]string
}

func newMergeToJSONTransform(all bool, args []string) *mergeToJSONTransform {
	ret := &mergeToJSONTransform{
		all:  all,
		keys: make(map[string]string),
	}

	for _, arg := range args {
		if !strings.Contains(arg, transformOperator) {
			ret.keys[arg] = ""
			continue
		}

		tokens := strings.Split(arg, transformOperator)
		ret.keys[tokens[0]] = tokens[1]
	}

	return ret
}

func (t *mergeToJSONTransform) Apply(pairs logfmt.Pairs) interface{} {
	obj := make(map[string]interface{})
	for _, pair := range pairs {
		typ, ok := t.keys[pair.Key]
		if !ok && !t.all {
			continue
		}

		switch typ {
		case "json":
			obj[pair.Key] = json.RawMessage(pair.Value)
		default:
			obj[pair.Key] = pair.Value
		}
	}

	data, _ := json.MarshalIndent(obj, "", "  ")

	return data
}

type dummyTransform struct{}

func (t *dummyTransform) Apply(pairs logfmt.Pairs) interface{} {
	return pairs
}

type transforms []transform

func (t transforms) Apply(pairs logfmt.Pairs) interface{} {
	ret := make(logfmt.Pairs, len(pairs))
	copy(ret, pairs)

	for _, tr := range t {
		tmp := tr.Apply(ret)
		if tmp == nil {
			return nil
		}

		v, ok := tmp.(logfmt.Pairs)
		if !ok {
			panic("can only chain transforms which produce pairs, not a merged result")
		}

		ret = v
	}

	return ret
}
