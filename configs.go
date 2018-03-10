package configs

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
)

func Load(filename string, cfg interface{}, options ...LoadOption) error {
	if file, err := os.Open(filename); err == nil {
		return Parse(file, cfg, options...)
	} else {
		return err
	}
}

func ParseBytes(data []byte, cfg interface{}, options ...LoadOption) error {
	return Parse(bytes.NewReader(data), cfg, options...)
}

func Parse(rdr io.Reader, cfg interface{}, options ...LoadOption) error {
	l := &configLoader{}

	for _, option := range options {
		option(l)
	}

	return l.Parse(rdr, cfg)
}

type configLoader struct {
	UseFlags *flag.FlagSet
}

func (l *configLoader) Parse(rdr io.Reader, cfg interface{}) error {
	return json.NewDecoder(rdr).Decode(cfg)
}
