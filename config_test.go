package ckoanf

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	Key         string `koanf:"key"`
	ABC         string `koanf:"abc"`
	NotSetField string `koanf:"not_set_field"`

	Nested Nested `koanf:"nested"`
}

type Nested struct {
	Foo string `koanf:"foo"`
}

// Non-function source, only used in testing - not exported.
func withRawSource[C ConfigModel](src Source) Option[C] {
	return func(mgr *Config[C]) error {
		mgr.sources = append(mgr.sources, src)
		return nil
	}
}

func (m TestModel) Validate() error {
	if m.Key == "" {
		return errors.New("Key is required")
	}
	if m.ABC != "" { // ABC must be length 3
		if len(m.ABC) != 3 {
			return errors.New("ABC must be length 3 if present")
		}
	}
	return nil
}

func TestConfig(t *testing.T) {
	t.Run("Load and Model", func(t *testing.T) {
		cfg, err := New(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("key = 'value'"), FileTypeTOML)))
		assert.NoError(t, err)

		err = cfg.Load()
		assert.NoError(t, err)

		model := cfg.Model()
		assert.Equal(t, "value", model.Key)
	})

	t.Run("Load with custom context", func(t *testing.T) {
		cfg, err := New(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("key = 'value'"), FileTypeTOML)))
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()

		err = cfg.Load(ctx)
		assert.NoError(t, err)

		err = cfg.Load(ctx, context.Background())
		assert.Error(t, err)
	})

	t.Run("Load and Model with invalid validation", func(t *testing.T) {
		invalidConfigOpt := WithSource(
			EmbeddedDefaults[*TestModel]([]byte("key = 'value'\nabc = 'abcd'"), FileTypeTOML),
		)

		cfg, err := New(&TestModel{}, invalidConfigOpt)
		assert.NoError(t, err)

		err = cfg.Load()
		assert.Error(t, err)

		// No error thrown f we disable validation
		cfgNoValidation, err := New(&TestModel{}, invalidConfigOpt, WithValidation[*TestModel](false))
		assert.NoError(t, err)

		err = cfgNoValidation.Load()
		assert.NoError(t, err)
	})

	t.Run("Init", func(t *testing.T) {
		cfg, err := Init(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("key = 'value'"), FileTypeTOML)))
		assert.NoError(t, err)

		model := cfg.Model()
		assert.Equal(t, "value", model.Key)
	})

	t.Run("Option errors", func(t *testing.T) {
		_, err := New(&TestModel{}, func(mgr *Config[*TestModel]) error {
			return errors.New("error")
		})
		assert.Error(t, err)

		_, err = Init(&TestModel{}, func(mgr *Config[*TestModel]) error {
			return errors.New("error")
		})
		assert.Error(t, err)
	})

	t.Run("Provider errors", func(t *testing.T) {
		// No error in load yet, it only errors when we try to load
		c, err := New(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("%"), FileTypeTOML)))
		assert.NoError(t, err)
		err = c.Load()
		assert.Error(t, err)

		_, err = Init(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("%"), FileTypeTOML)))
		assert.Error(t, err)
	})

	t.Run("Source with invalid type", func(t *testing.T) {
		myInvalidSource := Source{
			Type: SourceType("invalid"),
			Load: func(ctx context.Context, k *koanf.Koanf) error {
				return nil
			},
		}

		_, err := Init(&TestModel{}, withRawSource[*TestModel](myInvalidSource))
		assert.Error(t, err)
	})
}

type Empty struct{}

func (m Empty) Validate() error {
	return nil
}

func TestUnmarshalableConfig(t *testing.T) {
	_, err := Init(Empty{}) // Non-pointer
	assert.Error(t, err)
}

func TestSet(t *testing.T) {
	cfg, err := Init(&TestModel{}, WithSource(EmbeddedDefaults[*TestModel]([]byte("key = 'value'"), FileTypeTOML)))
	assert.NoError(t, err)

	err = cfg.Set("key", "new_value")
	assert.NoError(t, err)

	model := cfg.Model()
	assert.Equal(t, "new_value", model.Key)

	err = cfg.Set("nested.foo", "new_foo")
	assert.NoError(t, err)

	model = cfg.Model()
	assert.Equal(t, "new_foo", model.Nested.Foo)
}
