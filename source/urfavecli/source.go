// Package urfavecli implements a koanf.Source that loads the config from urfave/cli/v3 flags.
package urfavecli

import (
	"context"
	"fmt"

	"github.com/gzuidhof/ckoanf"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"
)

// Flags is a source that loads the config from urfave/cli/v3 flags.
func Flags[C ckoanf.ConfigModel](c *cli.Command) ckoanf.SourceFunc[C] {
	return func(mgr *ckoanf.Config[C]) (ckoanf.Source, error) {
		src := ckoanf.Source{
			Type: ckoanf.SourceTypePFlag,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				kprovider := newProvider(c, k.Delim())
				err := k.Load(kprovider, nil)
				if err != nil {
					return fmt.Errorf("failed to load config from urfave/cli/v3 flags: %w", err)
				}
				return nil
			},
		}
		return src, nil
	}
}
