# ckoanf

> Composable koanf

This package is a wrapper around [`koanf`](https://github.com/knadh/koanf) which allows for
* Composition of config sources in a specific order through composition.
* Defining initial values (for easy testing).
* Validation of config values (e.g. using the `ozzo-validation` package).
* Strict typing - config is stored in a strictly typed struct (no `interface{}` or `any`).

This package does not mess with any globals.

## Usage


```go
package main 

import (
    _ "embed"
    "github.com/gzuidhof/ckoanf"
)

// You define the struct equivalent of you config file to allow for strict typing
type AppConfig struct {
    Port int `koanf:"port"`
    Name string `koanf:"name"`
}

// The config type must have a `Validate` method.
// I recommend you use the `github.com/go-ozzo/ozzo-validation` package.
func (c AppConfig) Validate() error {
    if c.Port <= 0 {
        return fmt.Errorf("port must be positive, but was %d", c.Port)
    }
    return nil
}

// go:embed config.toml
var myEmbeddedConfigFile []byte

func InitConfig() *ckoanf.Config[*AppConfig]{
    configModel := &AppConfig{} // You can specify initial values in this struct (useful for testing!)
    c, err := ckoanf.Init(configModel
        ckoanf.WithSource(
            ckoanf.EmbeddedDefaults[*AppConfig](myEmbeddedConfigFile, ckoanf.FileTypeTOML), // Use the embedded file for default values.
            ckoanf.LocalFile[*AppConfig]("path/to/config.toml"),
            ckoanf.Env[*AppConfig]("MY_PREFIX_"), // Only loads env vars with this specific prefix
        ),
    )
    if err != nil {
        panic("error initializing config: " + err.Error())
    }
    // `c.K` can be used to access the Koanf instance, or you can use `c.Model()` to get the config model itself.
    return c
}
```

## Defaults
* A delimiter of `.` is used (as is the default for `koanf`).
* Environment varialbes are mapped such that a double underscore (`__`) becomes delimiter `.`.

## Non-Goals
For now this package does not support dynamic loading of additional configs.

## Testing
Test coverage is between 90% and 100%, let's try to keep it there.

# License
[MIT](./LICENSE).