package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/arnokay/arnobot-shared/pkg/assert"
)

const (
	ENV_MB_URL               = "MB_URL"
	ENV_DB_DSN               = "DB_DSN"
	ENV_TWITCH_WH_SECRET     = "TWITCH_WH_SECRET"
	ENV_BASE_URL             = "BASE_URL"
	ENV_PORT                 = "PORT"
	ENV_TWITCH_CLIENT_ID     = "TWITCH_CLIENT_ID"
	ENV_TWITCH_CLIENT_SECRET = "TWITCH_CLIENT_SECRET"
)

type config struct {
	Global   GlobalConfig
	Twitch   TwitchConfig
	MB       MBConfig
	DB       DBConfig
	Webhooks Webhooks
}

type TwitchConfig struct {
	ClientID     string
	ClientSecret string
}

type DBConfig struct {
	DSN          string
	MaxIdleConns int
	MaxOpenConns int
	MaxIdleTime  string
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
	Config = &config{
		Global: GlobalConfig{
			LogLevel: -4,
		},
	}

	if os.Getenv(ENV_PORT) != "" {
		port, err := strconv.Atoi(os.Getenv(ENV_PORT))
		assert.NoError(err, fmt.Sprintf("%v: not a number", ENV_PORT))
		Config.Global.Port = port
	}

	flag.StringVar(&Config.MB.URL, "mb-url", os.Getenv(ENV_MB_URL), "Message Broker URL")
	flag.IntVar(&Config.Global.LogLevel, "log-level", Config.Global.LogLevel, "Minimal Log Level (default: -4)")
	flag.StringVar(&Config.Webhooks.Secret, "wh-secret", os.Getenv(ENV_TWITCH_WH_SECRET), "secret for subscribing to webhooks")
	flag.StringVar(&Config.Global.BaseURL, "base-url", os.Getenv(ENV_BASE_URL), "public url")
	flag.IntVar(&Config.Global.Port, "port", Config.Global.Port, "http port")
	flag.StringVar(&Config.Twitch.ClientID, "t-client-id", os.Getenv(ENV_TWITCH_CLIENT_ID), "twitch client id")
	flag.StringVar(&Config.Twitch.ClientSecret, "t-client-secret", os.Getenv(ENV_TWITCH_CLIENT_SECRET), "twitch client id")
	flag.StringVar(&Config.Webhooks.Secret, "t-wh-client-secret", os.Getenv(ENV_TWITCH_CLIENT_SECRET), "twitch secret")
	flag.StringVar(&Config.DB.DSN, "db-dsn", os.Getenv(ENV_DB_DSN), "DB DSN")
	flag.IntVar(&Config.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.IntVar(&Config.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.StringVar(&Config.DB.MaxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	return Config
}
