package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/arnokay/arnobot-shared/applog"
	sharedMiddleware "github.com/arnokay/arnobot-shared/middlewares"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	sharedService "github.com/arnokay/arnobot-shared/service"
	"github.com/arnokay/arnobot-shared/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	apiController "github.com/arnokay/arnobot-twitch/internal/api/controller"
	apiMiddleware "github.com/arnokay/arnobot-twitch/internal/api/middleware"
	"github.com/arnokay/arnobot-twitch/internal/config"
	mbController "github.com/arnokay/arnobot-twitch/internal/mb/controller"
	"github.com/arnokay/arnobot-twitch/internal/service"
)

const AppName = "twitch"

type application struct {
	logger *slog.Logger

	msgBroker *nats.Conn
	api       *echo.Echo
	db        *pgxpool.Pool
	storage   storage.Storager

	apiControllers *apiController.Contollers
	apiMiddlewares *apiMiddleware.Middlewares

	mbControllers *mbController.Controllers

	services *service.Services
}

func main() {
	var app application

	// load config
	cfg := config.Load()

	// load logger
	logger := applog.Init(AppName, os.Stdout, cfg.Global.LogLevel)
	app.logger = logger

	// load db
	dbConn := openDB()
	app.db = dbConn
	app.storage = storage.NewStorage(app.db)

	// load message broker
	mbConn := openMB()
	app.msgBroker = mbConn

	// load services
	services := &service.Services{}
	services.TransactionService = sharedService.NewPgxTransactionService(app.db)
	services.AuthModule = sharedService.NewAuthModule(app.msgBroker)
	services.PlatformModule = sharedService.NewPlatformModuleOut(app.msgBroker)
	services.HelixManager = service.NewHelixManager(
		services.AuthModule,
		config.Config.Twitch.ClientID,
		config.Config.Twitch.ClientSecret,
	)
	services.TwitchService = service.NewTwitchService(services.HelixManager)
	services.WebhookService = service.NewWebhookService(services.HelixManager, services.TwitchService)
	services.BotService = service.NewBotService(
		app.storage,
		services.TransactionService,
		services.AuthModule,
		services.WebhookService,
		services.TwitchService,
	)
	app.services = services

	// load api middlewares
	app.apiMiddlewares = apiMiddleware.New(
		sharedMiddleware.NewAuthMiddleware(app.services.AuthModule),
	)

	// load api controllers
	app.apiControllers = &apiController.Contollers{
		WebhookController: apiController.NewWebhookController(
			app.apiMiddlewares,
			app.services.BotService,
			app.services.PlatformModule,
		),
	}

	// load mb controllers
	app.mbControllers = &mbController.Controllers{
		ChatController: mbController.NewChatController(app.services.TwitchService),
		BotController:  mbController.NewBotController(app.services.BotService),
	}

	app.Start()
}

func openDB() *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(config.Config.DB.DSN)
	assert.NoError(err, "openDB: cannot open database connection")

	cfg.MaxConns = int32(config.Config.DB.MaxOpenConns)
	cfg.MinConns = int32(config.Config.DB.MaxIdleConns)

	duration, err := time.ParseDuration(config.Config.DB.MaxIdleTime)
	assert.NoError(err, "openDB: cannot parse duration")

	cfg.MaxConnIdleTime = duration

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	assert.NoError(err, "openDB: cannot open database connection")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	assert.NoError(err, "openDB: cannot ping")

	return pool
}

func openMB() *nats.Conn {
	nc, err := nats.Connect(config.Config.MB.URL)
	assert.NoError(err, "openMB: cannot open message broker connection")

	return nc
}
