package controller

import (
	"encoding/json"
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

type WebhookController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	twitchService  *service.TwitchService
	botService     *service.BotService
	platformModule *sharedService.PlatformModuleOut
}

func NewWebhookController(
	middlewares *middleware.Middlewares,
	botService *service.BotService,
	platformModule *sharedService.PlatformModuleOut,
) *WebhookController {
	logger := applog.NewServiceLogger("ChatController")

	return &WebhookController{
		logger: logger,

		middlewares:    middlewares,
		botService:     botService,
		platformModule: platformModule,
	}
}

func (c *WebhookController) Routes(parentGroup *echo.Group) {
	parentGroup.POST("/callback", c.Callback, c.middlewares.VerifyTwitchWebhook)
}

func (c *WebhookController) Callback(ctx echo.Context) error {
	var rawEvent struct {
		Subscription helix.EventSubSubscription `json:"subscription"`
		Event        json.RawMessage            `json:"event"`
	}
	err := ctx.Bind(&rawEvent)
	if err != nil {
		c.logger.ErrorContext(ctx.Request().Context(), "cannot parse body", "event", "channel.chat.message", "err", err)
		return nil
	}

	switch ctx.Request().Header.Get("Twitch-Eventsub-Subscription-Type") {
	case helix.EventSubTypeChannelChatMessage:
		var event helix.EventSubChannelChatMessageEvent
		json.Unmarshal(rawEvent.Event, &event)

		bot, err := c.botService.SelectedBotGetByBroadcasterID(ctx.Request().Context(), event.BroadcasterUserID)
		if err != nil {
			c.logger.ErrorContext(ctx.Request().Context(), "cannot get selected bot")
			return nil
		}

		internalEvent := events.Message{
			EventCommon: events.EventCommon{
				Platform:      platform.Twitch,
				UserID:        bot.UserID,
				BroadcasterID: event.BroadcasterUserID,
				BotID:         bot.BotID,
			},
			MessageID: event.MessageID,
			// weird \U000e0000 appears in every second message
			Message:          strings.Replace(event.Message.Text, "\U000e0000", "", 1),
			ReplyTo:          event.Reply.ParentMessageID,
			BroadcasterLogin: event.BroadcasterUserLogin,
			BroadcasterName:  event.BroadcasterUserName,
			ChatterID:        event.ChatterUserID,
			ChatterName:      event.ChatterUserName,
			ChatterLogin:     event.ChatterUserLogin,
		}

		err = c.platformModule.ChatMessageNotify(ctx.Request().Context(), internalEvent)
		if err != nil {
			c.logger.ErrorContext(ctx.Request().Context(), "cannot send message to core")
			return nil
		}
	}

	return nil
}
