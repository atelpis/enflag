# Enflag [![Go Reference](https://pkg.go.dev/badge/github.com/atelpis/enflag.svg)](https://pkg.go.dev/github.com/atelpis/enflag) [![Go Report Card](https://goreportcard.com/badge/github.com/atelpis/enflag)](https://goreportcard.com/report/github.com/atelpis/enflag) [![codecov](https://codecov.io/gh/atelpis/enflag/graph/badge.svg?token=MH84VQP6EG)](https://codecov.io/gh/atelpis/enflag)

`Enflag` is a zero-dependency Golang package that simplifies configuring
applications via environment variables and command-line flags.

```bash
go get -u github.com/atelpis/enflag
```

## Why enflag?

The Go ecosystem has many configuration libraries, yet Enflag was created to
address a specific gap. Most Go configuration libraries fall into two extremes:

- Single-source libraries limit flexibility by supporting only env vars,
  flags, or config files.
- All-in-one solutions rely on reflection and struct tags, which can lead to
  hidden runtime errors, require memorizing library-specific syntax, and make
  debugging harder. A small typo in a struct tag or an unrecognized field may
  silently result in a zero value, often requiring extra attention to ensure
  correctness.

**Type safety through generics**: Enflag uses Go's generics to provide true
compile-time type checking.

**Balanced functionality:** Enflag embraces the simplicity of environment
variables and command-line flags,
covering the vast majority of use cases without bloated features. Its clean,
explicit API makes configuration predictable, readable, and easy to manage in
containerized environments.

## Features

- **Type-safe**: Generics-based implementation, no reflection
- **Docker-optimized**: Unified API for both environment variables and cmd flags
- **Comprehensive support**: Built-in parsing for all widely used types
- **Flexible**: Supports configurable slices, binary decoders,
  and JSON values
- **Extensible**: Easily add custom parsers
- **Minimalistic**: Clean, straightforward API
- **Lightweight**: Zero external dependencies

## Usage

The `Var` function takes a pointer to a configuration variable and assigns its
value according to the specified command-line flag or environment variable
using the `Bind` method.
Both sources are optional, but a flag has higher priority.

Additional methods like `WithDefault`, `WithTimeLayout`, etc. could be chained.

Behind the scenes, flags are handled by the standard library's
[flag.CommandLine flag set](https://pkg.go.dev/flag#CommandLine), meaning
you get the same help-message output and error handling. `Enflag` uses
generics to provide a cleaner and more convenient interface.

[See the full runnable example](https://pkg.go.dev/github.com/atelpis/enflag#example-package)

```go
type MyServiceConf struct {
    BaseURL *url.URL
    DBHost  string
    Dates  []time.Time
}

func main() {
    var conf MyServiceConf

    // Basic usage
    enflag.Var(&conf.BaseURL).Bind("BASE_URL", "base-url")

    // Simple bindings can be defined using the less verbose BindVar shortcut
    enflag.BindVar(&conf.BaseURL, "BASE_URL", "base-url")

    // Add settings
    enflag.Var(&conf.DBHost).
        WithDefault("127.0.0.1").
        WithFlagUsage("db hostname").
        Bind("DB_HOST", "db-host")

    // Slice
    enflag.Var(&conf.Dates).
        WithSliceSeparator("|").       // Split the slice using a non-default separator
        WithTimeLayout(time.DateOnly). // Use a non-default time layout
        BindEnv("DATES")               // Bind only the env variable, ignore the flag

    enflag.Parse()
}
```

Enflag supports the most essential data types out of the box like binary, strings,
numbers, time, URLs, IP and corresponding slices.
You can also use `VarFunc` with a custom parser to work with other types:

[See the full runnable example](https://pkg.go.dev/github.com/atelpis/enflag#example-package)

```go
func main() {
    var customVar int64

    parser := func(s string) (int64, error) {
        res, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            return 0, err
        }

        return res * 10, nil
    }
    enflag.VarFunc(&customVar, parser).Bind("CUSTOM", "custom")

    enflag.Parse()
}
```

## What about YAML?

While numerous packages handle complex configurations using YAML, TOML, JSON,
etc., `Enflag` focuses strictly on supporting container-oriented applications
with a reasonable configuration scope. It prioritizes simplicity and zero
dependency, making it a solid choice for microservices and cloud-native deployments. For more complex configuration needs, consider using a dedicated configuration management tool.
