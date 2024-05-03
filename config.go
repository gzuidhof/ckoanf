// Package ckoanf implements a config manager that uses koanf as its backend, the c stands for composable.
package ckoanf

import (
	"context"
	"fmt"
	"time"

	"github.com/knadh/koanf/v2"
)

// ConfigModel is the interface a config model must implement to be used with the config manager.
// We want to be able to validate the config model after it is unmarshalled from a config source.
type ConfigModel interface {
	Validate() error
}

// Config is a config manager that uses koanf as its backend, with a few extra features such as
// * A source system allowing you to compose multiple sources into a single config.
// * Validation.
// * A fixed config model (which can be retrieved using `Model()`).
type Config[C ConfigModel] struct {
	K     *koanf.Koanf
	model C

	// The sources to load the config from.
	// The order of the sources is important, as the config will be loaded in the same order.
	// Later sources will override values from earlier sources.
	sources []Source

	// Defaults to true
	validationEnabled bool
	strictMerge       bool
	loadTimeout       time.Duration
}

// New creates a new config manager.
// Note: this does not call `Load`. You may want to use `Init` instead.
func New[C ConfigModel](c C, opts ...Option[C]) (*Config[C], error) {
	mgr := &Config[C]{
		validationEnabled: true,
		strictMerge:       false,
		model:             c,
		loadTimeout:       time.Second * 10,
	}

	for i, opt := range opts {
		err := opt(mgr)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option %d: %w", i, err)
		}
	}

	for i, src := range mgr.sources {
		if err := src.Type.Valid(); err != nil {
			return nil, fmt.Errorf("invalid source type for provider %d: %w", i, err)
		}
	}

	mgr.K = koanf.NewWithConf(koanf.Conf{
		Delim:       defaultDelimiter,
		StrictMerge: mgr.strictMerge,
	})

	return mgr, nil
}

// Validate the config model by calling its `Validate` method.
func (mgr *Config[C]) Validate() error {
	if err := mgr.model.Validate(); err != nil {
		return fmt.Errorf("failed to validate config model: %w", err)
	}
	return nil
}

// Load the config from the given sources.
// Optionally takes a context to use for the load operation.
func (mgr *Config[C]) Load(ctxs ...context.Context) error {
	baseContext := context.Background()
	if len(ctxs) > 1 {
		return fmt.Errorf("too many contexts given")
	}
	if len(ctxs) == 1 {
		baseContext = ctxs[0]
	}

	ctx, cancel := context.WithTimeout(baseContext, mgr.loadTimeout)
	defer cancel()

	for i, source := range mgr.sources {
		if err := source.Load(ctx, mgr.K); err != nil {
			return fmt.Errorf("failed to load config from provider %d (type=%s): %w", i, source.Type, err)
		}
	}

	if err := mgr.K.Unmarshal("", mgr.model); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if mgr.validationEnabled {
		if err := mgr.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Init creates a new config manager and loads the config from the given sources.
// This is equivalent to calling `New` and `Load` in sequence.
func Init[C ConfigModel](c C, opts ...Option[C]) (*Config[C], error) {
	cfg, err := New(c, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create new config manager: %w", err)
	}

	if err := cfg.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return cfg, nil
}

// Model returns the underlying model of the config manager.
//
//nolint:ireturn,nolintlint // Generic return, false positive by the linter
func (mgr *Config[C]) Model() C {
	return mgr.model
}

// Set changes a value in the config by path key.
// Note that this requires unmarshaling and is fairly expensive.
func (mgr *Config[C]) Set(key string, value interface{}) error {
	err := mgr.K.Set(key, value)
	if err != nil {
		return fmt.Errorf("ckoanf failed to set value: %w", err)
	}

	err = mgr.K.Unmarshal("", mgr.model)
	if err != nil {
		return fmt.Errorf("ckoanf failed to unmarshal config: %w", err)
	}
	return nil
}
