package ckoanf

import (
	"fmt"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MyExampleConfig struct {
	Port    int                    `koanf:"port"`
	Address MyExampleConfigAddress `koanf:"address"`
}

type MyExampleConfigAddress struct {
	// This field must be exactly 2 characters long.
	CountryCode string `koanf:"country_code"`

	// This field is optional.
	City string `koanf:"city"`

	Language string `koanf:"language"`
}

func (c MyExampleConfig) Validate() error {
	fmt.Print(c.Address)
	return validation.ValidateStruct(&c,
		validation.Field(&c.Port, validation.Required, validation.Min(0), validation.Max(65535)),
		validation.Field(&c.Address),
	)
}

func (c MyExampleConfigAddress) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.CountryCode, validation.Required, validation.Length(2, 2)),
		validation.Field(&c.City, validation.Length(0, 100)),
		validation.Field(&c.Language, validation.Required, validation.Length(2, 2)),
	)
}

func TestMerging(t *testing.T) {
	conf1 := []byte(`
port = 8080

[address]
  country_code = "NL"
  city = "Amsterdam"
`)

	conf2 := []byte(`
[address]
  country_code = "INVALID"
`)

	conf3 := []byte(`
[address]
  country_code = "NL"
  city = ""
  language = "nl"
`)

	model := &MyExampleConfig{}

	c, err := New(
		model,
		WithSource(
			EmbeddedDefaults[*MyExampleConfig](conf1, FileTypeTOML),
			EmbeddedDefaults[*MyExampleConfig](conf2, FileTypeTOML),
			EmbeddedDefaults[*MyExampleConfig](conf3, FileTypeTOML),
		),
	)
	require.NoError(t, err)

	err = c.Load()
	require.NoError(t, err)

	assert.Equal(t, 8080, c.K.Int("port"))
	assert.Equal(t, "NL", c.K.String("address.country_code"))
	assert.Equal(t, "", c.K.String("address.city"))
	assert.Equal(t, "nl", c.K.String("address.language"))

	assert.Equal(t, 8080, model.Port)
	assert.Equal(t, "NL", model.Address.CountryCode)
	assert.Equal(t, "", model.Address.City)
	assert.Equal(t, "nl", model.Address.Language)

	assert.Equal(t, model, c.Model())
}
