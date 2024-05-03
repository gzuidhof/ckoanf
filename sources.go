package ckoanf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type SourceFunc[C ConfigModel] func(*Config[C]) (Source, error)

// OptionalSource wraps a source and makes it optional to load.
// If initialization of the source fails it will still error.
//
// If the source fails to load, but the error is allowed, the source will be skipped.
// If no allowed errors are provided, all errors are allowed.
func OptionalSource[C ConfigModel](src SourceFunc[C], allowedErrors ...error) SourceFunc[C] {
	isAllowedError := func(err error) bool {
		if len(allowedErrors) == 0 {
			return true
		}
		for _, allowed := range allowedErrors {
			if errors.Is(err, allowed) {
				return true
			}
		}
		return false
	}

	return func(mgr *Config[C]) (Source, error) {
		if src == nil {
			return Source{}, fmt.Errorf("source cannot be nil")
		}

		innerSrc, err := src(mgr)
		if err != nil {
			return Source{}, err
		}
		src := Source{
			Type: innerSrc.Type,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				err := innerSrc.Load(ctx, k)
				if err != nil && !isAllowedError(err) {
					return err
				}
				return nil
			},
		}

		return src, nil
	}
}

// EmbeddedDefaults is a source that loads the config from an embedded config file.
func EmbeddedDefaults[C ConfigModel](b []byte, filetype ConfigFileType) SourceFunc[C] {
	return func(mgr *Config[C]) (Source, error) {
		if err := filetype.Valid(); err != nil {
			return Source{}, err
		}
		parser := filetype.Parser()
		kprovider := rawbytes.Provider(b)

		src := Source{
			Type: SourceTypeDefault,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				err := k.Load(kprovider, parser)
				if err != nil {
					return fmt.Errorf("failed to load config from embedded defaults: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}

// LocalFile is a source that loads the config from a local file.
// The filetype is inferred from the file extension.
func LocalFile[C ConfigModel](filepath string) SourceFunc[C] {
	filetype := inferConfigFiletype(filepath)

	return func(mgr *Config[C]) (Source, error) {
		parser := filetype.Parser()
		kprovider := file.Provider(filepath)

		src := Source{
			Type: SourceTypeLocalFile,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				err := k.Load(kprovider, parser)
				if err != nil {
					return fmt.Errorf("failed to load config from local file: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}

// Env is a source that loads the config from environment variables.
//
// Only environment variables with the given prefix will be loaded.
func Env[C ConfigModel](prefix string) SourceFunc[C] {
	return func(mgr *Config[C]) (Source, error) {
		kprovider := env.Provider(prefix, defaultDelimiter, func(s string) string {
			// replace `__` with `.`, for example `PARENT__CHILD__NAME`
			// will be merged into the config as nested "parent.child.name"
			ret := strings.TrimPrefix(s, prefix)
			ret = strings.ReplaceAll(strings.ToLower(ret), "__", defaultDelimiter)
			return ret
		})

		src := Source{
			Type: SourceTypeEnv,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				err := k.Load(kprovider, nil)
				if err != nil {
					return fmt.Errorf("failed to load config from env vars: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}

// PFlags is a source that loads the config from posix command line flags.
func PFlags[C ConfigModel](flagset *pflag.FlagSet) SourceFunc[C] {
	return func(mgr *Config[C]) (Source, error) {
		if flagset == nil {
			return Source{}, fmt.Errorf("flagset cannot be nil")
		}

		src := Source{
			Type: SourceTypePFlag,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				kprovider := posflag.Provider(flagset, defaultDelimiter, k)
				err := k.Load(kprovider, nil)
				if err != nil {
					return fmt.Errorf("failed to load config from posix flags: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}

// Struct is a source that loads the config from a struct.
func Struct[C ConfigModel](s ConfigModel) SourceFunc[C] {
	if s == nil {
		return func(mgr *Config[C]) (Source, error) {
			return Source{}, fmt.Errorf("struct cannot be nil")
		}
	}

	return func(mgr *Config[C]) (Source, error) {
		kprovider := structs.Provider(s, "koanf")

		src := Source{
			Type: SourceTypeStruct,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				if err := k.Load(kprovider, nil); err != nil {
					return fmt.Errorf("failed to load config from struct: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}
