package ckoanf

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

//nolint:gochecknoglobals // This is a test fixture
var tomlFixture = `
key = "value"
abc = "def"
extra = "extra"

[nested]
  foo = "bar"
`

func TestEmbeddedDefaults(t *testing.T) {
	cfg, err := Init(&TestModel{},
		WithSource(
			EmbeddedDefaults[*TestModel]([]byte(tomlFixture), FileTypeTOML),
		))
	assert.NoError(t, err)

	assert.Equal(t, "value", cfg.Model().Key)
	assert.Equal(t, "def", cfg.Model().ABC)
	assert.Equal(t, "bar", cfg.Model().Nested.Foo)

	assert.Equal(t, "extra", cfg.K.String("extra"))
	assert.Equal(t, cfg.K.Exists("extra"), true)
	assert.False(t, cfg.K.Exists("not_set_field"))

	// Invalid file type should error
	_, err = Init(&TestModel{}, WithSource(
		EmbeddedDefaults[*TestModel]([]byte(tomlFixture), "invalid"),
	))
	assert.Error(t, err)
}

func TestLocalFile(t *testing.T) {
	myDir := t.TempDir()
	filepath := myDir + "/config.toml"
	model := &TestModel{}

	contents := []byte(tomlFixture)
	err := os.WriteFile(filepath, contents, 0o600)
	assert.NoError(t, err)

	cfg, err := Init(model, WithSource(
		LocalFile[*TestModel](filepath),
	))
	assert.NoError(t, err)

	assert.Equal(t, "value", cfg.Model().Key)
	assert.Equal(t, "def", cfg.Model().ABC)
	assert.Equal(t, "bar", cfg.Model().Nested.Foo)

	assert.Equal(t, "extra", cfg.K.String("extra"))
	assert.Equal(t, cfg.K.Exists("extra"), true)
	assert.False(t, cfg.K.Exists("not_set_field"))

	// Non-existent file should error
	_, err = Init(model, WithSource(LocalFile[*TestModel]("/non/existent/file.toml")))
	assert.Error(t, err)
}

func TestEnv(t *testing.T) {
	model := &TestModel{}

	prefix := "MY_PREFIX__"

	assert.NoError(t, os.Setenv(prefix+"KEY", "value"))
	assert.NoError(t, os.Setenv(prefix+"ABC", "def"))
	assert.NoError(t, os.Setenv(prefix+"EXTRA", "extra"))
	assert.NoError(t, os.Setenv(prefix+"NESTED__FOO", "bar"))

	cfg, err := Init(model, WithSource(Env[*TestModel](prefix)))
	assert.NoError(t, err)

	assert.Equal(t, "value", cfg.Model().Key)
	assert.Equal(t, "def", cfg.Model().ABC)
	assert.Equal(t, "bar", cfg.Model().Nested.Foo)

	assert.Equal(t, "extra", cfg.K.String("extra"))
	assert.Equal(t, cfg.K.Exists("extra"), true)
	assert.False(t, cfg.K.Exists("not_set_field"))
}

func TestFlags(t *testing.T) {
	model := &TestModel{}

	pflags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	pflags.String("key", "default", "")
	pflags.String("abc", "def", "Some usage explanation string")
	pflags.String("extra", "extra", "")
	pflags.String("nested.foo", "", "")

	err := pflags.Parse([]string{
		"--key", "value",
		"--nested.foo", "bar",
	})
	assert.NoError(t, err)

	cfg, err := Init(model, WithSource(PFlags[*TestModel](pflags)))
	assert.NoError(t, err)

	assert.Equal(t, "value", cfg.Model().Key)
	assert.Equal(t, "def", cfg.Model().ABC)
	assert.Equal(t, "bar", cfg.Model().Nested.Foo)
	assert.Equal(t, "extra", cfg.K.String("extra"))

	// Nil PFlags should error
	_, err = Init(model, WithSource(PFlags[*TestModel](nil)))
	assert.Error(t, err)
}

func TestOptionalSource(t *testing.T) {
	model := &Empty{}

	myError := errors.New("some error")

	_, err := Init(model, WithSource(OptionalSource[*Empty](nil)))
	assert.Error(t, err)

	myErrorSource := func(*Config[*Empty]) (Source, error) {
		return Source{
			Type: SourceTypeDefault,
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				return myError
			},
		}, nil
	}

	// All errors are ignored
	_, err = Init(model, WithSource(OptionalSource[*Empty](myErrorSource)))
	assert.NoError(t, err)

	// Here we add the error to the allowed errors list, so it should not error
	_, err = Init(model, WithSource(OptionalSource[*Empty](myErrorSource, myError)))
	assert.NoError(t, err)

	// Here we only allow a different error
	_, err = Init(model, WithSource(OptionalSource[*Empty](myErrorSource, errors.New("some other error"))))
	assert.Error(t, err)

	// Here the wrapped source returns an error on initialization, not on load. That should still error.
	_, err = New(model, WithSource(
		OptionalSource[*Empty](
			EmbeddedDefaults[*Empty](nil, ConfigFileType("invalid"))),
	))
	assert.Error(t, err)
}

func TestStruct(t *testing.T) {
	model := &TestModel{
		Key: "value",
		Nested: Nested{
			Foo: "bar",
		},
	}

	cfg, err := Init(model, WithSource(Struct[*TestModel](model)))
	assert.NoError(t, err)

	assert.Equal(t, "value", cfg.Model().Key)
	assert.Equal(t, "bar", cfg.Model().Nested.Foo)

	// Nil struct should error
	_, err = Init(model, WithSource(Struct[*TestModel](nil)))
	assert.Error(t, err)
}
