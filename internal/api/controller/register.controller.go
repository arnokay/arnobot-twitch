package controller

import (
	"log/slog"

	"arnobot-shared/appctx"
	"arnobot-shared/applog"
	"arnobot-shared/data"
	sharedService "arnobot-shared/service"
	"github.com/labstack/echo/v4"

	"arnobot-twitch/internal/api/middleware"
	"arnobot-twitch/internal/service"
)

type RegisterController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	webhookService    *service.WebhookService
	botService        *service.BotService
	authModuleService *sharedService.AuthModuleService
}

func NewRegisterController(
	middlewares *middleware.Middlewares,
	webhookService *service.WebhookService,
	botService *service.BotService,
	authModuleService *sharedService.AuthModuleService,
) *RegisterController {
	logger := applog.NewServiceLogger("register-controller")

	return &RegisterController{
		logger: logger,

		middlewares: middlewares,

		webhookService:    webhookService,
		botService:        botService,
		authModuleService: authModuleService,
	}
}

func (c *RegisterController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/register", c.middlewares.AuthMiddlewares.UserSessionGuard)
	g.POST("/", c.Register)
}

func (c *RegisterController) Register(ctx echo.Context) error {
	user := appctx.GetUser(ctx.Request().Context())

	userProvider, err := c.authModuleService.AuthProviderGet(ctx.Request().Context(), data.AuthProviderGet{
		UserID:   &user.ID,
		Provider: "twitch",
	})
	if err != nil {
		return err
	}

	bot, err := c.botService.SelectedBotGet(ctx.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	err = c.webhookService.SubscribeChannelChatMessageBot(ctx.Request().Context(), bot.TwitchUserID, userProvider.ProviderUserID)
	if err != nil {
		return err
	}

	return nil
}
