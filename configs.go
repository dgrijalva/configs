package configs

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
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
	Reader         io.Reader
	Args           []string
	flagMap        map[string]interface{}
}

func (l *configLoader) Parse(cfg interface{}) error {
	// TODO: flag parsing
	if l.UseFlags != nil {
		if err := l.ParseFlags(cfg); err != nil {
			return err
		}
	}

	// Parse json config
	if err := json.NewDecoder(l.Reader).Decode(cfg); err != nil {
		return err
	}

	return nil
}

func (l *configLoader) ParseFlags(cfg interface{}) error {
	l.GenerateFlags()

	// Figure out what args to use
	// Priority order is:
	// 1) Args supplied by WithArgs option
	// 2) Args supplied at the command line
	// 3) Empty arg list
	var args = l.Args
	if args == nil {
		if len(os.Args) > 1 {
			args = os.Args[1:]
		} else {
			args = []string{}
		}
	}

	if err := l.UseFlags.Parse(args); err != nil {
		return err
	}

	return nil
}

// Generate flags for config options
// Use JSON tags unless config tag is present
// TODO: look for congig tag
func (l *configLoader) GenerateFlags() error {

	return nil
}

type flagExtract struct {
	name string
	typ  reflect.Type
}

func (l *configLoader) extractFlagConfigs(ch chan flagExtract, cfg interface{}) {
	var v = reflect.ValueOf(cfg)
	fmt.Println(v)
}
