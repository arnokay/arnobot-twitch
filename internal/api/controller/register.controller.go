package controller

import (
	"fmt"
	"log/slog"

	"arnobot-shared/appctx"
	"arnobot-shared/applog"
	"arnobot-shared/data"
	"arnobot-shared/pkg/errs"
	sharedService "arnobot-shared/service"
	"github.com/labstack/echo/v4"

	"arnobot-twitch/internal/api/middleware"
	"arnobot-twitch/internal/service"
)

type RegisterController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	webhookService     *service.WebhookService
	botService         *service.BotService
	authModuleService  *sharedService.AuthModuleService
	transactionService sharedService.ITransactionService
}

func NewRegisterController(
	middlewares *middleware.Middlewares,
	webhookService *service.WebhookService,
	botService *service.BotService,
	authModuleService *sharedService.AuthModuleService,
	transactionService sharedService.ITransactionService,
) *RegisterController {
	logger := applog.NewServiceLogger("register-controller")

	return &RegisterController{
		logger: logger,

		middlewares: middlewares,

		webhookService:     webhookService,
		botService:         botService,
		authModuleService:  authModuleService,
		transactionService: transactionService,
	}
}

func (c *RegisterController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/register", c.middlewares.AuthMiddlewares.UserSessionGuard)
	g.POST("", c.Register)
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

	txCtx, err := c.transactionService.Begin(ctx.Request().Context())
	defer c.transactionService.Rollback(txCtx)
	if err != nil {
		return err
	}

	selectedBot, err := c.botService.SelectedBotGet(txCtx, user.ID)
	if err != nil {
		fmt.Println("kek")
		if !errs.IsAppErr(err) {
			fmt.Println("kek2")
			return err
		}

		var bot *data.TwitchBot

		bots, err := c.botService.BotsGet(txCtx, data.TwitchBotsGet{
			UserID: &user.ID,
		})
		if err != nil {
			return err
		}

		if len(bots) != 0 {
			bot = &bots[0]
		} else {
			defaultBot, err := c.botService.DefaultBotGet(txCtx)
			if err != nil {
				return err
			}
			bot, err = c.botService.BotCreate(txCtx, data.TwitchBotCreate{
				UserID:        user.ID,
				BotID:         defaultBot.BotID,
				BroadcasterID: userProvider.ProviderUserID,
			})
			if err != nil {
				return err
			}
		}

		selectedBot, err = c.botService.SelectedBotChange(txCtx, *bot)
		if err != nil {
			return err
		}
	}
	err = c.transactionService.Commit(txCtx)
	if err != nil {
		return err
	}

	err = c.webhookService.SubscribeChannelChatMessageBot(ctx.Request().Context(), selectedBot.BotID, userProvider.ProviderUserID)
	if err != nil {
		return err
	}

	return nil
}
