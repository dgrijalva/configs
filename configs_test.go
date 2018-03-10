package configs_test

import (
	"encoding/json"
	"testing"
)

type A struct {
	String string  `json:"string"`
	Float  float64 `json:"float"`
	Int    int64   `json:"int"`
}

var tests = []struct {
	name    string
	config  interface{}
	args    string
	err     error
	options []configs.LoadOption
}{
	{
		name:   "basic",
		config: &A{"foo", 1.23, 123},
	},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		// Write test config JSON to buffer
		buf := bytes.NewBuffer
		json.NewEncoder(buf).Encode(test.config)

		// Create new item to hold parsed results

		// Load config
		if test.options == nil {
			test.options = []configs.LoadOption{}
		}
		err := configs.Parse(buf, res, test.options...)
		if err != nil {

		}
	}
}
