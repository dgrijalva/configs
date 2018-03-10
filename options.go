package configs

import (
	"bytes"
	"flag"
	"io"
	"os"
)

type LoadOption func(*configLoader) error

// Call this method to enable flag parsing. Flags will
// be defined automatically to match config.
//
// If fs is not nil, fs will be used as the starting point.
// Otherwise, starting point will be an empty flag.FlagSet.
func WithFlags(fs *flag.FlagSet) LoadOption {
	if fs == nil {
		fs = flag.NewFlagSet("config", flag.ContinueOnError)
	}
	return func(l *configLoader) error {
		l.UseFlags = fs
		return nil
	}
}

// Load the config from a specified file path
func WithFile(filename string) LoadOption {
	return func(l *configLoader) error {
		if file, err := os.Open(filename); err == nil {
			l.Reader = file
		} else {
			return err
		}
		return nil
	}
}

// Load the config from json in []byte
func WithBytes(data []byte) LoadOption {
	return func(l *configLoader) error {
		l.Reader = bytes.NewReader(data)
		return nil
	}
}

// Load the config from an io.Reader
func WithReader(rdr io.Reader) LoadOption {
	return func(l *configLoader) error {
		l.Reader = rdr
		return nil
	}
}
