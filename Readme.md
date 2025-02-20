# Enflag [![Go Reference](https://pkg.go.dev/badge/github.com/atelpis/enflag.svg)](https://pkg.go.dev/github.com/atelpis/enflag) [![Go Report Card](https://goreportcard.com/badge/github.com/atelpis/enflag)](https://goreportcard.com/report/github.com/atelpis/enflag) [![codecov](https://codecov.io/gh/atelpis/enflag/graph/badge.svg?token=MH84VQP6EG)](https://codecov.io/gh/atelpis/enflag)

`Enflag` is a zero-dependency Golang package that simplifies configuring
applications via environment variables and command-line flags.

```bash
go get -u github.com/atelpis/enflag
```

Go provides a fantastic [flag package](https://pkg.go.dev/flag),
which does a great job of working with the command line. However, the common
practice for cloud-oriented applications is to support both flags and
corresponding environment variables. `Enflag` addresses this by defining
both configuration sources in a single function call.

## Usage

The `Bind` function takes a pointer to a configuration variable and assigns its
value according to the specified command-line flag or environment variable.
Both sources are optional, but a flag has higher priority.

Behind the scenes, flags are handled by the standard library's
[flag.CommandLine flag set](https://pkg.go.dev/flag#CommandLine), meaning
you get the same help-message output and error handling. `Enflag` uses
generics to provide a cleaner and more convenient interface.

[See the extended runnable example](https://pkg.go.dev/github.com/atelpis/enflag#example-package)

```go
type MyServiceConf struct {
    DBHost  string
    DBPort  int
    BaseURL *url.URL
}

func main() {
    var conf MyServiceConf

    enflag.Bind(&conf.DBHost, "DB_HOST", "db-host", "127.0.0.1", "db hostname")
    enflag.Bind(&conf.DBPort, "DB_PORT", "db-port", 5432, "db port")
    enflag.Bind(&conf.BaseURL, "BASE_URL", "base-url", nil, "website base url")

    enflag.Parse()
}
```

`Bind` supports the most essential datatypes out of the box like strings,
numbers, duration, URLs, and IP. You can also use `BindFunc` with a custom
parser to work with other types:

[See the extended runnable example](https://pkg.go.dev/github.com/atelpis/enflag#example-package)

```go
func main() {
    var timeVar *time.Time

   enflag.BindFunc(
        &timeVar,
        "ITIME",
        "itime",
        nil,
        "important time in unix-milli format",
        func(ts string) (*time.Time, error) {
            ms, err := strconv.ParseInt(ts, 10, 64)
            if err != nil {
                return nil, err
            }

            t := time.UnixMilli(ms)
            return &t, nil
        },
    )

    enflag.Parse()
}
```

## What about YAML?

While numerous packages handle complex configurations using YAML, TOML, JSON,
etc., `Enflag` focuses strictly on supporting container-oriented applications
with a reasonable configuration scope. It prioritizes simplicity and zero
dependency, making it a solid choice for microservices and cloud-native deployments. For more complex configuration needs, consider using a dedicated configuration management tool.
