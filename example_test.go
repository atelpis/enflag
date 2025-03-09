package enflag_test

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/atelpis/enflag"
)

func Example() {
	// emulate environment variables
	{
		os.Setenv("ENV", "develop")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("SECRET", "AQID")
	}

	var conf MyServiceConf

	// Both env and flag are defined and provided,
	// flag value will be used as it has higher priority.
	enflag.Var(&conf.DBHost).WithDefault("127.0.0.1").WithFlagUsage("db hostname").Bind("DB_HOST", "db-host")

	// Both env and flag are defined, but neither is provided.
	// The value of DBPort will default to 5432.
	enflag.Var(&conf.DBPort).WithDefault(5432).Bind("DB_PORT", "db-port")

	// Example of parsing a non-primitive type.
	enflag.Var(&conf.BaseURL).Bind("BASE_URL", "base-url")

	// Both env and flag sources are optional. Skip the flag definition,
	// and retrieve this value only from the environment.
	enflag.Var(&conf.Env).BindEnv("ENV")

	// By default binary variables as parsed as base64 string.
	enflag.Var(&conf.Secret).Bind("SECRET", "secret")

	// Custom parser
	{
		parser := func(s string) (int64, error) {
			res, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0, err
			}

			return res * 10, nil
		}
		enflag.VarFunc(&conf.CustomVar, parser).Bind("CUSTOM", "custom")
	}

	// emulate flag values
	{
		flag.CommandLine.Set("db-host", "db.mysrv.int")
		flag.CommandLine.Set("base-url", "https://my-website.com")
		flag.CommandLine.Set("custom", "3")
	}

	enflag.Parse()

	RunMyService(&conf)
}

func RunMyService(c *MyServiceConf) error {
	fmt.Println("Starting service:")

	fmt.Printf("- Env: %s\n", c.Env)
	fmt.Printf("- DB Host: %s\n", c.DBHost)
	fmt.Printf("- DB Port: %d\n", c.DBPort)
	fmt.Printf("- Base URL: %v\n", c.BaseURL)
	fmt.Printf("- Custom var: %v\n", c.CustomVar)
	fmt.Printf("- Secret len: %d\n", len(c.Secret))

	return nil
}

type MyServiceConf struct {
	Env string

	DBHost  string
	DBPort  int
	BaseURL *url.URL
	Secret  []byte

	CustomVar int64
}
