package ckoanf

import "fmt"

const defaultDelimiter = "."

// Option is a functional option that can be used to configure the config manager.
type Option[C ConfigModel] func(*Config[C]) error

// WithValidation enables or disables validation of the config.
func WithValidation[C ConfigModel](v bool) Option[C] {
	return func(mgr *Config[C]) error {
		mgr.validationEnabled = v
		return nil
	}
}

// WithSource adds one or more sources to the config manager.
//
// The order of the sources is important, as the config will be loaded in the same order.
// Later sources will override values from earlier sources.
func WithSource[C ConfigModel](srcs ...SourceFunc[C]) Option[C] {
	return func(mgr *Config[C]) error {
		for i, src := range srcs {
			s, err := src(mgr)
			if err != nil {
				return fmt.Errorf("failed to create source %d: %w", i, err)
			}
			mgr.sources = append(mgr.sources, s)
		}
		return nil
	}
}
