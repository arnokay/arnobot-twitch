package controller

import (
	"log/slog"
	"strings"

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/events"
	"github.com/arnokay/arnobot-shared/platform"
	sharedService "github.com/arnokay/arnobot-shared/service"
	"github.com/labstack/echo/v4"
	"github.com/nicklaw5/helix/v2"

	"github.com/arnokay/arnobot-twitch/internal/api/middleware"
	"github.com/arnokay/arnobot-twitch/internal/service"
)

type ChannelWebhookController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	twitchService  *service.TwitchService
	botService     *service.BotService
	platformModule *sharedService.PlatformModuleOut
}

func NewChatController(
	middlewares *middleware.Middlewares,
	botService *service.BotService,
	platformModule *sharedService.PlatformModuleOut,
) *ChannelWebhookController {
	logger := applog.NewServiceLogger("ChatController")

	return &ChannelWebhookController{
		logger: logger,

		middlewares:    middlewares,
		botService:     botService,
		platformModule: platformModule,
	}
}

func (c *ChannelWebhookController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/callback", c.middlewares.VerifyTwitchWebhook)
	g.POST("/channel-chat-message", c.ChannelChatMessageHandler)
}

func (c *ChannelWebhookController) ChannelChatMessageHandler(ctx echo.Context) error {
	var event struct {
		Subscription helix.EventSubSubscription            `json:"subscription"`
		Event        helix.EventSubChannelChatMessageEvent `json:"event"`
	}
	err := ctx.Bind(&event)
	if err != nil {
		c.logger.ErrorContext(ctx.Request().Context(), "cannot parse body", "event", "channel.chat.message", "err", err)
		return nil
	}

	bot, err := c.botService.SelectedBotGetByBroadcasterID(ctx.Request().Context(), event.Event.BroadcasterUserID)
	if err != nil {
		c.logger.ErrorContext(ctx.Request().Context(), "cannot get selected bot")
		return nil
	}

	internalEvent := events.Message{
		EventCommon: events.EventCommon{
			Platform:      platform.Twitch,
			BroadcasterID: event.Event.BroadcasterUserID,
			BotID:         bot.BotID,
		},
		MessageID:        event.Event.MessageID,
		Message:          strings.Replace(event.Event.Message.Text, "\U000e0000", "", 1),
		ReplyTo:          event.Event.Reply.ParentMessageID,
		BroadcasterLogin: event.Event.BroadcasterUserLogin,
		BroadcasterName:  event.Event.BroadcasterUserName,
		ChatterID:        event.Event.ChatterUserID,
		ChatterName:      event.Event.ChatterUserName,
		ChatterLogin:     event.Event.ChatterUserLogin,
	}

	err = c.platformModule.ChatMessageNotify(ctx.Request().Context(), internalEvent)
	if err != nil {
		c.logger.ErrorContext(ctx.Request().Context(), "cannot send message to core")
		return nil
	}

	return nil
}
