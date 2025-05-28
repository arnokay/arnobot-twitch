package controller

import (
	"fmt"
	"log/slog"

	"arnobot-shared/applog"
	"arnobot-shared/pkg/errs"
	"github.com/labstack/echo/v4"
	"github.com/nicklaw5/helix/v2"

	"arnobot-twitch/internal/api/middleware"
)

type ChannelWebhookController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares
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
	var event helix.EventSubChannelChatMessageEvent
	err := ctx.Bind(&event)
	if err != nil {
		c.logger.Error("cannot parse body", "event", "channel.chat.message", "err", err)
		return errs.ErrInvalidInput
	}

	fmt.Println(event)

	return nil
}
