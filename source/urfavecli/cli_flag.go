package urfavecli

import (
	"fmt"

	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"
)

var _ koanf.Provider = (*provider)(nil)

// provider for `urfave/cli/v3` flags.
type provider struct {
	delimiter string
	ctx       *cli.Command
}

func newProvider(ctx *cli.Command, delimiter string) *provider {
	return &provider{
		ctx:       ctx,
		delimiter: delimiter,
	}
}

func (p *provider) Read() (map[string]interface{}, error) {
	mp := make(map[string]interface{})

	for _, name := range p.ctx.FlagNames() {
		mp[name] = p.ctx.Value(name)
	}

	return maps.Unflatten(mp, p.delimiter), nil
}

// ReadBytes is not supported by the env koanf.
func (p *provider) ReadBytes() ([]byte, error) {
	return nil, fmt.Errorf("not supported")
}
