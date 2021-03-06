package configs_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dgrijalva/configs"
	"reflect"
	"strings"
	"testing"
)

type A struct {
	String string  `json:"string"`
	Float  float64 `json:"float"`
	Int    int64   `json:"int"`
}

type B struct {
	Foo string `json:"foo"`
}

type C struct {
	Foo *B `json:"foo"`
}

type D struct {
	*A
	Foo string `json:"foo"`
}

type E struct {
	Foo *A     `json:"foo"`
	Bar string `json:"bar"`
}

var tests = []struct {
	name    string
	config  interface{}
	expect  interface{}
	args    string
	err     error
	options []configs.LoadOption
}{
	{
		name:   "basic",
		config: &A{"foo", 1.23, 123},
	},
	{
		name:   "flags",
		expect: &A{"bar", 4.56, 456},
		args:   "-string bar -float 4.56 -int 456",
	},
	{
		name:   "flag overwrite",
		config: &A{"foo", 1.23, 123},
		args:   "-string bar",
		expect: &A{"bar", 1.23, 123},
	},
	{
		name:   "nested",
		config: &C{&B{Foo: "bar"}},
	},
	{
		name:   "nested w/ flags",
		config: &C{},
		expect: &C{&B{Foo: "bar"}},
		args:   "-foo.foo bar",
	},
	{
		name:   "embedded",
		config: &D{&A{String: "str"}, "bar"},
	},
	{
		name:   "embedded",
		config: &D{Foo: "bar"},
		expect: &D{A: &A{String: "foo"}, Foo: "bar"},
		args:   "-string foo",
	},
	{
		name:    "config file",
		options: []configs.LoadOption{configs.WithFileFlag("config")},
		expect:  &E{Foo: &A{Int: 123, String: "foo"}, Bar: "baz"},
		args:    "-config test/e_001.json",
	},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		if test.expect == nil && test.config != nil {
			test.expect = test.config
		}

		// Create new item to hold parsed results
		var res interface{} = reflect.New(reflect.Indirect(reflect.ValueOf(test.expect)).Type()).Interface()

		// Load config
		testOptions := []configs.LoadOption{}
		if test.config != nil {
			// Write test config JSON to buffer
			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(test.config)
			testOptions = append(testOptions, configs.WithReader(buf))
		}
		if test.args != "" {
			// Automatically create a flagset that's not the default one
			// This will be overwritten if WithFlags is used in the test data
			testOptions = append(testOptions, configs.WithFlags(flag.NewFlagSet(fmt.Sprintf("test[%v]", test.name), flag.ContinueOnError)))
			// Convert arg string to WithArgs option
			testOptions = append(testOptions, configs.WithArgs(strings.Split(test.args, " ")))
		}
		if test.options != nil {
			for _, opt := range test.options {
				testOptions = append(testOptions, opt)
			}
		}
		err := configs.Load(res, testOptions...)

		// Handle error cases
		if err != nil {
			if test.err == nil {
				t.Errorf("[%v] Unexpected error: %v", test.name, err)
			} else if err != test.err {
				t.Errorf("[%v] Error did not meet expectations. Expected %v got %v", test.name, test.err, err)

			}
			continue
		}

		// Handle success cases
		if test.expect == nil {
			test.expect = test.config
		}
		if !reflect.DeepEqual(test.expect, res) {
			t.Errorf("[%v] Parsed config didn't match expectation. Expected %T %+v got %T %+v", test.name, test.expect, debugJSON(test.expect), res, debugJSON(res))
		}
	}
}

func debugJSON(x interface{}) string {
	stuff, _ := json.Marshal(x)
	return string(stuff)
}
