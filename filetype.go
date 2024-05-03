package ckoanf

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
)

// ConfigFileType represents the type of the config file.
type ConfigFileType string

const (
	FileTypeYAML ConfigFileType = "yaml"
	FileTypeTOML ConfigFileType = "toml"
	FileTypeJSON ConfigFileType = "json"
)

// Infer the config file type from the file path. If the file path does not
// have a valid extension, the default file type is returned.
//
// If no default file type is provided, the default is TOML.
func inferConfigFiletype(path string, defaultFiletype ...ConfigFileType) ConfigFileType {
	if strings.HasSuffix(path, ".toml") {
		return FileTypeTOML
	} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return FileTypeYAML
	} else if strings.HasSuffix(path, ".json") {
		return FileTypeJSON
	}

	if len(defaultFiletype) > 0 {
		return defaultFiletype[0]
	}
	return FileTypeTOML
}

// String returns the string representation of the config file type.
func (c ConfigFileType) String() string {
	return string(c)
}

// Valid checks if the config file type is valid.
func (c ConfigFileType) Valid() error {
	switch c {
	case FileTypeYAML, FileTypeTOML, FileTypeJSON:
		return nil
	default:
		return fmt.Errorf("invalid config file type: %s", c)
	}
}

// Parser returns the parser for the config file type.
// Panics if the config file type is invalid (use `Valid()` first to check).
// nolint:ireturn,nolintlint // This is required to return this interface to be a valid Source.
func (c ConfigFileType) Parser() koanf.Parser {
	switch c {
	case FileTypeYAML:
		return yaml.Parser()
	case FileTypeTOML:
		return toml.Parser()
	case FileTypeJSON:
		return json.Parser()
	default:
		panic(fmt.Errorf("invalid config file type: %s", c))
	}
}
