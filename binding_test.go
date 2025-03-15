package enflag

import (
	"encoding/hex"
	"flag"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestBind(t *testing.T) {
	ErrorHandlerFunc = OnErrorIgnore

	type tc struct {
		name string

		f     func(*testing.T) []func()
		envs  []string
		flags []string
	}

	cases := []tc{
		{
			name: "Default value",

			// no values provided, use default
			envs: nil, flags: nil,
			f: func(t *testing.T) []func() {
				var target string
				Var(&target).WithDefault("localhost").Bind("HOST", "host")

				return toSlice(func() { checkVal(t, "localhost", target) })
			},
		},
		{
			name:  "Read env",
			envs:  []string{"PORT", "443"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target int
				Var(&target).WithDefault(80).Bind("PORT", "port")

				return toSlice(func() { checkVal(t, int(443), target) })
			},
		},
		{
			name:  "Flag has priority",
			envs:  []string{"PORT", "8080"},
			flags: []string{"port", "443"},
			f: func(t *testing.T) []func() {
				var target int
				Var(&target).WithDefault(80).Bind("PORT", "port")

				return toSlice(func() { checkVal(t, int(443), target) })
			},
		},
		{
			name:  "BindVar",
			envs:  []string{"PORT", "443"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target int
				BindVar(&target, "PORT", "port", "port helper")

				return toSlice(func() { checkVal(t, int(443), target) })
			},
		},
		{
			name: "String slice",
			envs: []string{"LABELS", "inbox,sent"},
			f: func(t *testing.T) []func() {
				var target []string
				Var(&target).BindEnv("LABELS")

				return toSlice(func() { checkSlice(t, []string{"inbox", "sent"}, target) })
			},
		},
		{
			name:  "Base64 bytes",
			envs:  []string{"SECRET", "AQID"},
			flags: []string{"secret-hex", "010203"},
			f: func(t *testing.T) []func() {
				var targetBase64 []byte
				Var(&targetBase64).BindEnv("SECRET")

				var targetHEX []byte
				Var(&targetHEX).WithStringDecodeFunc(hex.DecodeString).BindFlag("secret-hex")

				return []func(){
					func() { checkSlice(t, []byte{1, 2, 3}, targetBase64) },
					func() { checkSlice(t, []byte{1, 2, 3}, targetHEX) },
				}
			},
		},
		{
			name:  "Int slice",
			envs:  []string{"IDS", "1,3,4"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []int
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []int{1, 3, 4}, target) })
			},
		},
		{
			name:  "Int64",
			envs:  []string{"PORT", "8888"},
			flags: []string{"port", "443"},
			f: func(t *testing.T) []func() {
				var target int64
				Var(&target).WithDefault(80).WithFlagUsage("int64 value").Bind("PORT", "port")

				return toSlice(func() { checkVal(t, int64(443), target) })
			},
		},
		{
			name:  "Int64 slice",
			envs:  []string{"IDS", "1,3,4"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []int64
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []int64{1, 3, 4}, target) })
			},
		},
		{
			name:  "Uint",
			envs:  []string{"PORT", "8888"},
			flags: []string{"port", "443"},
			f: func(t *testing.T) []func() {
				var target uint
				Var(&target).WithDefault(80).WithFlagUsage("uint value").Bind("PORT", "port")

				return toSlice(func() { checkVal(t, uint(443), target) })
			},
		},
		{
			name:  "Uint slice",
			envs:  []string{"IDS", "1,3,4"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []uint
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []uint{1, 3, 4}, target) })
			},
		},
		{
			name:  "Uint64",
			envs:  []string{"PORT", "8888"},
			flags: []string{"port", "443"},
			f: func(t *testing.T) []func() {
				var target uint64
				Var(&target).WithDefault(80).WithFlagUsage("uint64 value").Bind("PORT", "port")

				return toSlice(func() { checkVal(t, uint64(443), target) })
			},
		},
		{
			name:  "Uint64 slice",
			envs:  []string{"IDS", "1,3,4"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []uint64
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []uint64{1, 3, 4}, target) })
			},
		},
		{
			name:  "Float64",
			envs:  []string{"LLM_TEMP", "0.35"},
			flags: []string{"llm-temp", "0.45"},
			f: func(t *testing.T) []func() {
				var target float64
				Var(&target).WithDefault(1).WithFlagUsage("llm requests temperature").Bind("LLM_TEMP", "llm-temp")

				return toSlice(func() { checkVal(t, float64(0.45), target) })
			},
		},
		{
			name:  "Float64 slice",
			envs:  []string{"IDS", "1,3,4"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []float64
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []float64{1, 3, 4}, target) })
			},
		},
		{
			name: "Boolean",
			envs: []string{
				"DEBUG", "0",
				"REQUIRE_2FA", "false",
			},
			flags: nil,
			f: func(t *testing.T) []func() {
				var targetNumeric bool
				var targetStr bool

				Var(&targetNumeric).WithDefault(true).BindEnv("DEBUG")
				Var(&targetStr).WithDefault(true).BindEnv("REQUIRE_2FA")

				return []func(){
					func() { checkVal(t, false, targetNumeric) },
					func() { checkVal(t, false, targetStr) },
				}
			},
		},
		{
			name:  "Bool slice",
			envs:  []string{"IDS", "1,true,false"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target []bool
				Var(&target).BindEnv("IDS")

				return toSlice(func() { checkSlice(t, []bool{true, true, false}, target) })
			},
		},
		{
			name: "JSON",
			envs: []string{"OBJ", `{"a": 1, "s": [1, 2, 3]}`},

			// for testing parsing from flags
			flags: []string{"obj", `{"a": 4, "s": [3, 2, 1]}`},
			f: func(t *testing.T) []func() {
				type obj struct {
					A int    `json:"a"`
					S []uint `json:"S"`
				}

				var targetEnv obj
				var targetFlag obj
				VarJSON(&targetEnv).BindEnv("OBJ")
				VarJSON(&targetFlag).BindFlag("obj")

				return []func(){
					func() { checkVal(t, 1, targetEnv.A) },
					func() { checkSlice(t, []uint{1, 2, 3}, targetEnv.S) },

					func() { checkVal(t, 4, targetFlag.A) },
					func() { checkSlice(t, []uint{3, 2, 1}, targetFlag.S) },
				}
			},
		},

		{
			name: "URL",
			// for testing parsing from env
			envs: []string{"BASE_ADMIN_URL", "https://admin.my-domain.com/home"},

			// for testing parsing from flags
			flags: []string{"base-url", "https://app.my-domain.com/home"},
			f: func(t *testing.T) []func() {
				var target url.URL
				var targetAdmin url.URL
				def := url.URL{Scheme: "http", Host: "localhost", Path: "/sign-in"}

				Var(&target).WithDefault(def).WithFlagUsage("application base url").BindFlag("base-url")
				Var(&targetAdmin).WithDefault(def).BindEnv("BASE_ADMIN_URL")

				return []func(){
					func() { checkVal(t, "https", target.Scheme) },
					func() { checkVal(t, "app.my-domain.com", target.Host) },
					func() { checkVal(t, "/home", target.Path) },

					func() { checkVal(t, "admin.my-domain.com", targetAdmin.Host) },
				}
			},
		},
		{
			name: "URL slice",
			envs: []string{"URLS", "https://app.my-domain.com/home;https://app.my-domain.com/dashboard"},

			f: func(t *testing.T) []func() {
				var target []url.URL
				Var(&target).WithSliceSeparator(";").BindEnv("URLS")

				return []func(){
					func() { checkVal(t, 2, len(target)) },
					func() { checkVal(t, "https", target[0].Scheme) },
					func() { checkVal(t, "app.my-domain.com", target[0].Host) },
					func() { checkVal(t, "/home", target[0].Path) },

					func() { checkVal(t, "https", target[1].Scheme) },
					func() { checkVal(t, "app.my-domain.com", target[1].Host) },
					func() { checkVal(t, "/dashboard", target[1].Path) },
				}
			},
		},
		{
			name: "URL pointer",
			// for testing parsing from env
			envs: []string{"BASE_ADMIN_URL", "https://admin.my-domain.com/home"},

			// for testing parsing from flags
			flags: []string{"base-url", "https://app.my-domain.com/home"},

			f: func(t *testing.T) []func() {
				var target *url.URL
				var targetAdmin *url.URL
				var targetNil *url.URL
				def := &url.URL{Scheme: "http", Host: "localhost", Path: "/sign-in"}

				Var(&target).WithDefault(def).BindFlag("base-url")
				Var(&targetAdmin).WithDefault(def).BindEnv("BASE_ADMIN_URL")
				Var(&targetNil).BindEnv("PROMO_URL")

				return []func(){
					func() { checkVal(t, "https", target.Scheme) },
					func() { checkVal(t, "app.my-domain.com", target.Host) },
					func() { checkVal(t, "/home", target.Path) },

					func() { checkVal(t, "admin.my-domain.com", targetAdmin.Host) },

					func() { checkVal(t, nil, targetNil) },
				}
			},
		},
		{
			name: "IP",

			// for testing parsing from env
			envs: []string{"DNS_IP", "127.0.0.8"},

			// for testing parsing from flags
			flags: []string{"balancer-ip", "10.56.2.138"},
			f: func(t *testing.T) []func() {
				var target net.IP
				var targetBalancer *net.IP
				def := net.IP{127, 0, 0, 1}

				Var(&target).WithDefault(def).BindEnv("DNS_IP")
				Var(&targetBalancer).WithFlagUsage("ip address of the balancer").BindFlag("balancer-ip")

				return []func(){
					func() { checkVal(t, "127.0.0.8", target.String()) },
					func() { checkVal(t, "10.56.2.138", targetBalancer.String()) },
				}
			},
		},
		{
			name: "IP slice",

			envs: []string{"DNS_IPS", "127.0.0.8,127.0.0.9"},

			f: func(t *testing.T) []func() {
				var target []net.IP

				Var(&target).BindEnv("DNS_IPS")

				return []func(){
					func() { checkVal(t, 2, len(target)) },
					func() { checkVal(t, "127.0.0.8", target[0].String()) },
					func() { checkVal(t, "127.0.0.9", target[1].String()) },
				}
			},
		},
		{
			name:  "Time",
			envs:  []string{"DATE_3339", "2025-03-07T12:34:56Z", "OPT_TIME", "2025-03-07T12:34:56Z"},
			flags: []string{"date", "2025-03-07", "date-range", "2025-01-01:2025-03-07"},
			f: func(t *testing.T) []func() {
				layoutDate := "2006-01-02"

				expect, _ := time.Parse(time.RFC3339, "2025-03-07T12:34:56Z")
				expectDateOnly := time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC)
				expectRange := []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC),
				}

				var target3339 time.Time
				var targetOpt *time.Time
				var targetStampMillis time.Time
				var targetRange []time.Time

				Var(&target3339).BindEnv("DATE_3339")
				Var(&targetOpt).BindEnv("OPT_TIME")
				Var(&targetStampMillis).WithTimeLayout(layoutDate).BindFlag("date")
				Var(&targetRange).WithSliceSeparator(":").WithTimeLayout(layoutDate).BindFlag("date-range")

				return []func(){
					func() { checkVal(t, expect, target3339) },
					func() { checkVal(t, expect, *targetOpt) },
					func() { checkVal(t, expectDateOnly, targetStampMillis) },
					func() { checkSlice(t, expectRange, targetRange) },
				}
			},
		},
		{
			name:  "Duration",
			envs:  []string{"TTL", "5m"},
			flags: []string{"ttls", "5s,1m,3m"},
			f: func(t *testing.T) []func() {
				var target time.Duration
				var targetSlice []time.Duration

				Var(&target).WithDefault(30*time.Second).Bind("TTL", "ttl")
				Var(&targetSlice).BindFlag("ttls")

				return []func(){
					func() { checkVal(t, 5*time.Minute, target) },
					func() { checkSlice(t, []time.Duration{5 * time.Second, time.Minute, 3 * time.Minute}, targetSlice) },
				}
			},
		},
		{
			name:  "Overwrite default with zero",
			envs:  []string{"ALERT_THRESHOLD", "0"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target int

				Var(&target).WithDefault(5).BindEnv("ALERT_THRESHOLD")

				return toSlice(func() { checkVal(t, int(0), target) })
			},
		},
		{
			name:  "Overwrite default with zero flag",
			envs:  []string{"ALERT_THRESHOLD", "10"},
			flags: []string{"alert-thr", "0"},
			f: func(t *testing.T) []func() {
				var intVar int

				Var(&intVar).WithDefault(5).Bind("ALERT_THRESHOLD", "alert-thr")

				return toSlice(func() { checkVal(t, int(0), intVar) })
			},
		},
		{
			name:  "Custom parser",
			envs:  []string{"MY_FORMAT", "aaa"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target string
				parser := func(s string) (string, error) {
					return s + "-bbb", nil
				}
				VarFunc(&target, parser).
					WithFlagUsage("my format help").
					WithDefault("a").
					Bind("MY_FORMAT", "my-format")

				return toSlice(func() { checkVal(t, "aaa-bbb", target) })
			},
		},

		// invalid data
		{
			name: "Uint bad env",
			envs: []string{"PORT", "4-4-3"},
			f: func(t *testing.T) []func() {
				var target uint

				Var(&target).WithDefault(80).Bind("PORT", "port")

				return toSlice(func() { checkVal(t, uint(80), target) })
			},
		},
		{
			name: "URL bad env",
			envs: []string{"BAD_ADMIN_URL", "123"},

			f: func(t *testing.T) []func() {
				var targetAdmin url.URL
				def := url.URL{}

				Var(&targetAdmin).WithDefault(def).BindEnv("BAD_ADMIN_URL")

				return toSlice(func() { checkVal(t, "", targetAdmin.Host) })
			},
		},
		{
			name: "Bad slice env",
			envs: []string{"PORTS", "one,two"},

			f: func(t *testing.T) []func() {
				var target []int

				Var(&target).BindEnv("PORTS")

				return toSlice(func() { checkSlice(t, []int{}, target) })
			},
		},
		{
			name:  "Bad slice flag",
			flags: []string{"ports", "one,two"},

			f: func(t *testing.T) []func() {
				var target []int

				Var(&target).BindFlag("ports")

				return toSlice(func() { checkSlice(t, []int{}, target) })
			},
		},

		{
			name: "IP bad env",
			envs: []string{"DNS_IP", "aaa-bbb"},

			f: func(t *testing.T) []func() {
				var target net.IP

				Var(&target).WithDefault(net.IP{}).BindEnv("DNS_IP")

				return toSlice(func() { checkVal(t, "<nil>", target.String()) })
			},
		},
		{
			name:  "Custom bad flag",
			flags: []string{"my-format", "aaa"},
			f: func(t *testing.T) []func() {
				var target int
				parser := func(s string) (int, error) {
					return strconv.Atoi(s)
				}
				VarFunc(&target, parser).WithDefault(10).Bind("MY_FORMAT", "my-format")

				return toSlice(func() { checkVal(t, 10, target) })
			},
		},
		{
			name:  "Deprecated Bind",
			envs:  []string{"PORT", "8080"},
			flags: []string{"port", "443"},
			f: func(t *testing.T) []func() {
				var target int
				Bind(&target, "PORT", "port", 80, "int value")

				return toSlice(func() { checkVal(t, int(443), target) })
			},
		},
		{
			name:  "Deprecated BindFunc",
			envs:  []string{"MY_FORMAT", "aaa"},
			flags: nil,
			f: func(t *testing.T) []func() {
				var target string
				BindFunc(&target, "MY_FORMAT", "my-format", "a", "int value", func(s string) (string, error) {
					return s + "-bbb", nil
				})

				return toSlice(func() { checkVal(t, "aaa-bbb", target) })
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reset()

			for _, pair := range toPairs(c.envs) {
				os.Setenv(pair[0], pair[1])
			}

			checks := c.f(t)
			for _, pair := range toPairs(c.flags) {
				flag.Set(pair[0], pair[1])
			}

			Parse()
			for _, checkF := range checks {
				checkF()
			}
		})
	}
}

func checkVal[A comparable](t *testing.T, want A, got A) {
	t.Helper()

	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func checkSlice[A comparable](t *testing.T, want []A, got []A) {
	t.Helper()

	if len(want) != len(got) {
		t.Errorf("expected %v, got %v", want, got)
		return
	}

	for i := range want {
		if want[i] != got[i] {
			t.Errorf("want %v, got %v, mismatch at pos %d: %v != %v", want, got, i, want[i], got[i])
			return
		}
	}
}

func reset() {
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func toPairs(s []string) [][2]string {
	res := make([][2]string, 0, len(s)/2)
	for i := range s {
		if i%2 == 0 {
			continue
		}
		res = append(res, [2]string{s[i-1], s[i]})
	}
	return res
}

func toSlice[T any](v T) []T {
	sl := make([]T, 1)
	sl[0] = v
	return sl
}
