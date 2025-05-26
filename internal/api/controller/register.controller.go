package controller

import (
	"log/slog"

	"arnobot-shared/appctx"
	"arnobot-shared/applog"
	"github.com/labstack/echo/v4"

	"arnobot-twitch/internal/api/middleware"
	"arnobot-twitch/internal/service"
)

type RegisterController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	webhookService *service.WebhookService
	botService     *service.BotService
}

func NewRegisterController(
	middlewares *middleware.Middlewares,
	webhookService *service.WebhookService,
	botService *service.BotService,
) *RegisterController {
	logger := applog.NewServiceLogger("register-controller")

	return &RegisterController{
		logger: logger,

		middlewares: middlewares,

		webhookService: webhookService,
    botService: botService,
	}
}

func (c *RegisterController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/register", c.middlewares.AuthMiddlewares.UserSessionGuard)
	g.POST("/", c.Register)
}

func (c *RegisterController) Register(ctx echo.Context) error {
	_ = appctx.GetUser(ctx.Request().Context())

	return nil
}
