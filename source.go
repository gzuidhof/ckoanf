package ckoanf

import (
	"context"
	"fmt"

	"github.com/knadh/koanf/v2"
)

// Source is a config source, which is something that can be loaded into a koanf config.
type Source struct {
	Type SourceType
	Load func(context.Context, *koanf.Koanf) error
}

type SourceType string

const (
	SourceTypeDefault   SourceType = "default"
	SourceTypeLocalFile SourceType = "file"
	SourceTypeEnv       SourceType = "env"
	SourceTypePFlag     SourceType = "pflag"
	SourceTypeStruct    SourceType = "struct"
)

func (p SourceType) String() string {
	return string(p)
}

func (p SourceType) Valid() error {
	switch p {
	case SourceTypeDefault, SourceTypeLocalFile, SourceTypeEnv, SourceTypePFlag, SourceTypeStruct:
		return nil
	default:
		return fmt.Errorf("invalid provider type: %s", p)
	}
}
