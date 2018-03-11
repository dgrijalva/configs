package configs

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
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
	flagMap        map[int]flagExtract
}

func (l *configLoader) Parse(cfg interface{}) error {
	// Extract flags from config struct. Parse flags.
	if l.UseFlags != nil {
		if err := l.ParseFlags(cfg); err != nil {
			return err
		}
	}

	// Parse json config
	if err := json.NewDecoder(l.Reader).Decode(cfg); err != nil {
		return err
	}

	// Apply flags to config struct
	if l.UseFlags != nil {
		if err := l.ApplyFlags(cfg); err != nil {
			return err
		}
	}

	return nil
}

func (l *configLoader) ParseFlags(cfg interface{}) error {
	if err := l.GenerateFlags(cfg); err != nil {
		return err
	}

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

func (l *configLoader) ApplyFlags(cfg interface{}) error {
	var v = reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Config was type %v should be a *struct", v.Kind())
	}
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Config was type %v should be a *struct", v.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		f := l.flagMap[i]
		switch f.typ.Kind() {
		case reflect.String:
			field.SetString(*f.value.(*string))
		case reflect.Float64:
			field.SetFloat(*f.value.(*float64))
		case reflect.Int64:
			field.SetInt(*f.value.(*int64))
		}
	}

	return nil
}

// Generate flags for config options
// Use JSON tags unless config tag is present
func (l *configLoader) GenerateFlags(cfg interface{}) error {
	var v = reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Config was type %v should be a *struct", v.Kind())
	}
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Config was type %v should be a *struct", v.Kind())
	}

	if l.flagMap == nil {
		l.flagMap = make(map[int]flagExtract)
	}

	ch := make(chan flagExtract, 16)
	go l.extractFlagConfigs(ch, v)
	for f := range ch {
		switch f.typ.Kind() {
		case reflect.String:
			f.value = l.UseFlags.String(f.name, "", f.desc)
		case reflect.Float64:
			f.value = l.UseFlags.Float64(f.name, 0, f.desc)
		case reflect.Int64:
			f.value = l.UseFlags.Int64(f.name, 0, f.desc)
		default:
			// Dunno what to do
			continue
		}
		l.flagMap[f.index] = f
	}

	return nil
}

type flagExtract struct {
	index int
	name  string
	desc  string
	typ   reflect.Type
	value interface{}
}

func (l *configLoader) extractFlagConfigs(ch chan flagExtract, cfg reflect.Value) {
	// Iterate over struct fields and extract flag identifiers
	typ := cfg.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fe := flagExtract{index: i, typ: field.Type}

		if name, ok := field.Tag.Lookup("config"); ok {
			fe.name = strings.Split(name, ",")[0]
		} else if name, ok := field.Tag.Lookup("json"); ok {
			fe.name = strings.Split(name, ",")[0]
		}

		ch <- fe
	}
	close(ch)
}
