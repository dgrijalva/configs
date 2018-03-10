package configs

import (
	"encoding/json"
	"flag"
	"io"
)

// Load config.
func Load(cfg interface{}, options ...LoadOption) error {
	l := &configLoader{}

	for _, option := range options {
		if err := option(l); err != nil {
			return err
		}
	}

	return l.Parse(cfg)
}

type configLoader struct {
	ConfigFileFlag *flag.Flag
	UseFlags       *flag.FlagSet
	io.Reader
}

func (l *configLoader) Parse(cfg interface{}) error {
	// TODO: flag parsing

	// Parse json config
	if err := json.NewDecoder(l.Reader).Decode(cfg); err != nil {
		return err
	}

	return nil
}
