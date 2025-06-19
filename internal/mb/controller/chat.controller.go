package controller

import (
	"fmt"
	"log/slog"

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/apptype"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/arnokay/arnobot-shared/topics"
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
	chatMessageSendTopic := topics.PlatformBroadcasterChatMessageSend.Build(platform.Twitch, topics.Any)
	_, err := conn.QueueSubscribe(
		chatMessageSendTopic,
		chatMessageSendTopic,
		c.ChatMessageSend,
	)
	assert.NoError(err, fmt.Sprintf("MBChatController cannot subscribe to the topic: %s", chatMessageSendTopic))
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
