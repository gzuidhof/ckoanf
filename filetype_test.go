package ckoanf

import (
	"testing"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
)

func TestConfigFileType(t *testing.T) {
	t.Run("inferConfigFiletype", func(t *testing.T) {
		assert.Equal(t, FileTypeTOML, inferConfigFiletype("config.toml"))
		assert.Equal(t, FileTypeYAML, inferConfigFiletype("config.yaml"))
		assert.Equal(t, FileTypeYAML, inferConfigFiletype("config.yml"))
		assert.Equal(t, FileTypeJSON, inferConfigFiletype("config.json"))
		assert.Equal(t, FileTypeTOML, inferConfigFiletype("config"))
		assert.Equal(t, FileTypeYAML, inferConfigFiletype("config", FileTypeYAML))
	})

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "toml", FileTypeTOML.String())
		assert.Equal(t, "yaml", FileTypeYAML.String())
	})

	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, FileTypeTOML.Valid())
		assert.NoError(t, FileTypeYAML.Valid())
		assert.NoError(t, FileTypeJSON.Valid())
		assert.Error(t, ConfigFileType("invalid").Valid())
		assert.Error(t, ConfigFileType("").Valid())
	})

	t.Run("Parser", func(t *testing.T) {
		assert.IsType(t, &toml.TOML{}, FileTypeTOML.Parser())
		assert.IsType(t, &yaml.YAML{}, FileTypeYAML.Parser())
		assert.IsType(t, &json.JSON{}, FileTypeJSON.Parser())
		assert.Panics(t, func() { ConfigFileType("invalid").Parser() })
	})
}
