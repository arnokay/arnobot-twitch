package main

import (
	"log/slog"
	"os"

	"arnobot-shared/applog"
	sharedMiddleware "arnobot-shared/middlewares"
	"arnobot-shared/pkg/assert"
	sharedService "arnobot-shared/service"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"

	apiController "arnobot-twitch/internal/api/controller"
	apiMiddleware "arnobot-twitch/internal/api/middleware"
	"arnobot-twitch/internal/config"
	"arnobot-twitch/internal/service"
)

const APP_NAME = "twitch-webhooks"

type application struct {
	logger *slog.Logger

	msgBroker *nats.Conn
	api       *echo.Echo

	apiControllers *apiController.Contollers
	apiMiddlewares *apiMiddleware.Middlewares

	services *service.Services
}

func main() {
	var app application

	// load config
	cfg := config.Load()

	// load logger
	logger := applog.Init(APP_NAME, os.Stdout, cfg.Global.LogLevel)
	app.logger = logger

	// load message broker
	mbConn := openMB()
	app.msgBroker = mbConn

	// load services
	app.services = &service.Services{
		AuthModuleService: sharedService.NewAuthModuleService(app.msgBroker),
		HelixManager: sharedService.NewHelixManager(
			app.services.AuthModuleService,
			config.Config.Twitch.ClientID,
			config.Config.Twitch.ClientSecret,
		),
	}

	// load api middlewares
	app.apiMiddlewares = apiMiddleware.New(
		sharedMiddleware.NewAuthMiddleware(app.services.AuthModuleService),
	)

	app.apiControllers = &apiController.Contollers{
		RegisterController: apiController.NewRegisterController(app.apiMiddlewares),
	}
	// app.Start()
}

func openMB() *nats.Conn {
	nc, err := nats.Connect(config.Config.MB.URL)
	assert.NoError(err, "openMB: cannot open message broker connection")

	return nc
}
