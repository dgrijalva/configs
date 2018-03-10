package configs

import (
	"bytes"
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

func ParseBytes(config []byte, cfg interface{}, options ...LoadOption) error {
	return Parse(bytes.NewReader(config), cfg, options...)
}

func Parse(config io.Reader, cfg interface{}, options ...LoadOption) error {
	l := &configLoader{
		UseFlags: false,
	}

	for _, option := range options {
		option(l)
	}

	return l.Parse(config, cfg)
}

type configLoader struct {
	UseFlags *flag.FlagSet
}

func (l *configLoader) Parse(config io.Reader, cfg interface{}) error {
	return nil
}
