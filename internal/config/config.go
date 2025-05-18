package config

import (
	"flag"
	"os"
)

const (
	ENV_DB_DSN    = "MB_URL"
	ENV_WH_SECRET = "WH_SECRET"
)

type config struct {
	Global   GlobalConfig
	MB       MBConfig
	Webhooks Webhooks
}

type GlobalConfig struct {
	LogLevel int
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

	flag.StringVar(&Config.MB.URL, "mb-url", os.Getenv(ENV_DB_DSN), "Message Broker URL")
	flag.IntVar(&Config.Global.LogLevel, "log-level", Config.Global.LogLevel, "Minimal Log Level (default: -4)")
	flag.StringVar(&Config.Webhooks.Secret, "wh-secret", os.Getenv(ENV_WH_SECRET), "secrett for subscribing to webhooks")

	flag.Parse()

	return Config
}
