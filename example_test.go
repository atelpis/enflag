package enflag_test

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/atelpis/enflag"
)

func Example() {
	// emulate environment variables
	{
		os.Setenv("ENV", "develop")
		os.Setenv("DB_HOST", "localhost")
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

	// Custom time parser
	{
		parser := func(ts string) (*time.Time, error) {
			ms, err := strconv.ParseInt(ts, 10, 64)
			if err != nil {
				return nil, err
			}

			t := time.UnixMilli(ms)
			return &t, nil
		}
		enflag.VarFunc(&conf.ImportantTime, parser).Bind("ITIME", "itime")
	}

	// emulate flag values
	{
		flag.CommandLine.Set("db-host", "db.mysrv.int")
		flag.CommandLine.Set("base-url", "https://my-website.com")
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

	return nil
}

type MyServiceConf struct {
	Env string

	DBHost  string
	DBPort  int
	BaseURL *url.URL

	ImportantTime *time.Time
}
