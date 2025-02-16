package enflag

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
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
	Bind(&conf.DBHost, "DB_HOST", "db-host", "127.0.0.1", "db hostname")

	// Both env and flag are defined, but neither is provided.
	// The value of DBPort will default to 5432.
	Bind(&conf.DBPort, "DB_PORT", "db-port", 5432, "db port")

	// Example of parsing a non-primitive type.
	Bind(&conf.BaseURL, "BASE_URL", "base-url", nil, "website base url")

	// Both env and flag sources are optional. Skip the flag definition,
	// and retrieve this value only from the environment.
	Bind(&conf.Env, "ENV", "", "local", "")

	// Custom time parser
	BindFunc(
		&conf.ImportantTime,
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

	// emulate flag values
	{
		flag.CommandLine.Set("db-host", "db.mysrv.int")
		flag.CommandLine.Set("base-url", "https://my-website.com")
	}

	Parse()

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
