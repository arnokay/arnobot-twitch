package controller

import (
	"fmt"
	"log/slog"

	"arnobot-shared/applog"
	"arnobot-shared/apptype"
	"arnobot-shared/pkg/assert"
	"arnobot-shared/platform"
	"arnobot-shared/topics"
	"github.com/nats-io/nats.go"

	"arnobot-twitch/internal/service"
)

type ChatController struct {
	twitchService *service.TwitchService

	logger *slog.Logger
}

func NewChatController(
	twitchService *service.TwitchService,
) *ChatController {
	logger := applog.NewServiceLogger("mb-chat-controller")

	return &ChatController{
		twitchService: twitchService,

		logger: logger,
	}
}

func (c *ChatController) Connect(conn *nats.Conn) {
	_, err := conn.QueueSubscribe(
		topics.PlatformChatMessageSend.Platform(platform.Twitch),
		topics.PlatformChatMessageSend.Platform(platform.Twitch),
		c.ChatMessageSend,
	)
	assert.NoError(err, fmt.Sprintf("MBChatController cannot subscribe to the topic: %s", topics.PlatformChatMessageSend.Platform(platform.Twitch)))
}

func (c *ChatController) ChatMessageSend(msg *nats.Msg) {
	var payload apptype.PlatformChatMessageSend

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.twitchService.AppSendChannelMessage(ctx, payload.Data.BotID, payload.Data.BroadcasterID, payload.Data.Message, payload.Data.ReplyTo)
	if err != nil {
		c.logger.ErrorContext(
			ctx,
			"cannot send message to channel",
			"payload", payload,
		)
		return
	}
}
