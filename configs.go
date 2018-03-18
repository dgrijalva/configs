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
	UseFlags     *flag.FlagSet
	Reader       io.Reader
	Args         []string
	fileFlagName string               // config file path will be specified by this flag
	flagMap      map[int]*flagExtract // what flags were defined and how do they map to config
	setFlags     map[string]byte      // set identifying which flags were set
}

type flagExtract struct {
	index    int
	name     string
	cmdName  string
	desc     string
	typ      reflect.Type
	value    interface{}
	children map[int]*flagExtract
}

func (l *configLoader) Parse(cfg interface{}) error {
	// Create the flag described by WithFileFlag
	if l.fileFlagName != "" {
		if l.UseFlags == nil {
			l.UseFlags = flag.CommandLine
		}
		l.UseFlags.String(l.fileFlagName, "", "path to config file")
	}

	// Extract flags from config struct. Parse flags.
	if l.UseFlags != nil {
		if err := l.ParseFlags(cfg); err != nil {
			return err
		}
		if l.fileFlagName != "" {
			nameFlag := l.UseFlags.Lookup(l.fileFlagName)
			WithFile(nameFlag.Value.String())(l)
		}
	}

	// Parse json config
	if l.Reader != nil {
		if err := json.NewDecoder(l.Reader).Decode(cfg); err != nil {
			return err
		}
	}

	// Apply flags to config struct
	if l.UseFlags != nil {
		if err := l.ApplyFlags(cfg, l.flagMap); err != nil {
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

	// Capture which flags have been set
	l.setFlags = make(map[string]byte)
	l.UseFlags.Visit(func(f *flag.Flag) {
		l.setFlags[f.Name] = 1
	})

	return nil
}

func (l *configLoader) ApplyFlags(cfg interface{}, flagMap map[int]*flagExtract) error {
	// Verify that cfg is a valid type. Indirect pointer to get to struct value.
	var v = reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Config was type %v should be a *struct, A", v.Type())
	}
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Config was type %v should be a *struct, B", v.Type())
	}

	// Iterate over struct fields and set the values from flagMap
	for i := 0; i < v.NumField(); i++ {
		var field = v.Field(i)
		var f = flagMap[i]
		var fieldFlagSet = (l.setFlags[f.cmdName] == 1)

		if f.children != nil {
			// For children, handle recursively. Create empty structs where necessary
			var value = field
			if !value.IsValid() {
				if field.Type().Kind() == reflect.Ptr {
					value = reflect.New(field.Type().Elem())
				} else {
					value = reflect.New(field.Type())
				}
				field.Set(value)
			}
			// Recurse both value and flagMap
			if err := l.ApplyFlags(value.Interface(), f.children); err != nil {
				return err
			}
		} else if fieldFlagSet {
			// If the field was set with a flag, overwrite existing value
			// Copy values from parsed flags into struct
			switch f.typ.Kind() {
			case reflect.String:
				field.SetString(*f.value.(*string))
			case reflect.Float64:
				field.SetFloat(*f.value.(*float64))
			case reflect.Int64:
				field.SetInt(*f.value.(*int64))
			}
		}
	}

	return nil
}

// Generate flags for config options
// Priority:
// 1) flag struct tag (name,description) (TODO: use '-' to skip generating flags for property)
// 2) json struct tag (see encoding/json documentation)
// 3) TODO: field name
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
		l.flagMap = make(map[int]*flagExtract)
	}

	l.extractFlagConfigs(v.Type(), l.flagMap)
	l.declareFlags("", l.flagMap)

	return nil
}

func (l *configLoader) declareFlags(parentName string, flagMap map[int]*flagExtract) {
	for _, f := range flagMap {
		f.cmdName = f.name
		if parentName != "" {
			f.cmdName = strings.Join([]string{parentName, f.cmdName}, ".")
		}

		if f.children != nil && len(f.children) > 0 {
			l.declareFlags(f.cmdName, f.children)
		} else {
			switch f.typ.Kind() {
			case reflect.String:
				f.value = l.UseFlags.String(f.cmdName, "", f.desc)
			case reflect.Float64:
				f.value = l.UseFlags.Float64(f.cmdName, 0, f.desc)
			case reflect.Int64:
				f.value = l.UseFlags.Int64(f.cmdName, 0, f.desc)
			default:
				// Dunno what to do
				continue
			}
		}
	}
}

// recursively traverse config struct type and identify possible flags
// nested values are separated by a .
func (l *configLoader) extractFlagConfigs(typ reflect.Type, parent map[int]*flagExtract) {
	// Iterate over struct fields and extract flag identifiers
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		fe := &flagExtract{index: i, typ: field.Type}

		if name, ok := field.Tag.Lookup("flag"); ok {
			fe.name = strings.Split(name, ",")[0]
		} else if name, ok := field.Tag.Lookup("json"); ok {
			fe.name = strings.Split(name, ",")[0]
		}

		if field.Type.Kind() == reflect.Struct {
			fe.children = make(map[int]*flagExtract)
			l.extractFlagConfigs(field.Type, fe.children)
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			fe.children = make(map[int]*flagExtract)
			l.extractFlagConfigs(field.Type.Elem(), fe.children)
		}
		parent[i] = fe
	}
}
