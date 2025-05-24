package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"arnobot-shared/pkg/assert"
)

const (
	ENV_DB_DSN    = "MB_URL"
	ENV_WH_SECRET = "WH_SECRET"
	ENV_BASE_URL  = "BASE_URL"
	ENV_PORT      = "PORT"
)

type config struct {
	Global   GlobalConfig
	MB       MBConfig
	Webhooks Webhooks
}

type GlobalConfig struct {
	LogLevel int
	BaseURL  string
	Port     int
}

type MBConfig struct {
	URL string
}

type Webhooks struct {
	Secret string
}

var Config *config

func Load() *config {
	Config = &config{}

	if os.Getenv(ENV_PORT) != "" {
		port, err := strconv.Atoi(os.Getenv(ENV_PORT))
		assert.NoError(err, fmt.Sprintf("%v: not a number", ENV_PORT))
		Config.Global.Port = port
	}

	flag.StringVar(&Config.MB.URL, "mb-url", os.Getenv(ENV_DB_DSN), "Message Broker URL")
	flag.IntVar(&Config.Global.LogLevel, "log-level", Config.Global.LogLevel, "Minimal Log Level (default: -4)")
	flag.StringVar(&Config.Webhooks.Secret, "wh-secret", os.Getenv(ENV_WH_SECRET), "secrett for subscribing to webhooks")
	flag.StringVar(&Config.Global.BaseURL, "base-url", os.Getenv(ENV_BASE_URL), "public url")
	flag.IntVar(&Config.Global.Port, "port", Config.Global.Port, "http port")

	flag.Parse()

	return Config
}
