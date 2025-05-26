package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"arnobot-shared/pkg/assert"
)

const (
	ENV_DB_DSN        = "MB_URL"
	ENV_WH_SECRET     = "WH_SECRET"
	ENV_BASE_URL      = "BASE_URL"
	ENV_PORT          = "PORT"
	ENV_CLIENT_ID     = "CLIENT_ID"
	ENV_CLIENT_SECRET = "SECRET"
)

type config struct {
	Global   GlobalConfig
	Twitch   TwitchConfig
	MB       MBConfig
	Webhooks Webhooks
}

type TwitchConfig struct {
	ClientID     string
	ClientSecret string
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
	flag.StringVar(&Config.Webhooks.Secret, "wh-secret", os.Getenv(ENV_WH_SECRET), "secret for subscribing to webhooks")
	flag.StringVar(&Config.Global.BaseURL, "base-url", os.Getenv(ENV_BASE_URL), "public url")
	flag.IntVar(&Config.Global.Port, "port", Config.Global.Port, "http port")
	flag.StringVar(&Config.Twitch.ClientID, "client-id", os.Getenv(ENV_CLIENT_ID), "twitch client id")
	flag.StringVar(&Config.Twitch.ClientSecret, "client-secret", os.Getenv(ENV_CLIENT_SECRET), "twitch secret")

	flag.Parse()

	return Config
}
