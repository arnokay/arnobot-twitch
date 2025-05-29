package controller

import (
	"fmt"
	"log/slog"

	"arnobot-shared/applog"
	"github.com/labstack/echo/v4"
	"github.com/nicklaw5/helix/v2"

	"arnobot-twitch/internal/api/middleware"
	"arnobot-twitch/internal/service"
)

type ChannelWebhookController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares

	twitchService *service.TwitchService
}

func NewChatController(middlewares *middleware.Middlewares) *ChannelWebhookController {
	logger := applog.NewServiceLogger("ChatController")

	return &ChannelWebhookController{
		logger: logger,

		middlewares: middlewares,
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
		c.logger.Error("cannot parse body", "event", "channel.chat.message", "err", err)
		return nil
	}

	fmt.Println(event.Event.Message.Text)


	return nil
}
