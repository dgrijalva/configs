package configs

import (
	"flag"
)

type LoadOption func(*configLoader)

// Call this method to enable flag parsing. Flags will
// be defined automatically to match config.
//
// If fs is not nil, fs will be used as the starting point.
// Otherwise, starting point will be an empty flag.FlagSet.
func WithFlags(fs *flag.FlagSet) LoadOption {
	if fs == nil {
		fs = flag.NewFlagSet("config", flag.ContinueOnError)
	}
	return func(l *configLoader) {
		l.UseFlags = fs
	}
}
