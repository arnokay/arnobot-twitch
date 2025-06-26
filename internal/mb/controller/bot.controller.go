package controller

import (
	"log/slog"

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/apptype"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/arnokay/arnobot-shared/topics"
	"github.com/nats-io/nats.go"

	"github.com/arnokay/arnobot-twitch/internal/service"
)

type BotController struct {
	botService *service.BotService

	logger *slog.Logger
}

func NewBotController(
	botService *service.BotService,
) *BotController {
	logger := applog.NewServiceLogger("mb-bot-controller")

	return &BotController{
		botService: botService,

		logger: logger,
	}
}

func (c *BotController) Connect(conn *nats.Conn) {
	topic := topics.TopicBuilder(topics.PlatformStartBot).Platform(platform.Twitch).Build()
	_, err := conn.QueueSubscribe(topic, topic, c.StartBot)
	assert.NoError(err, "cannot subscribe to: "+topic)
	topic = topics.TopicBuilder(topics.PlatformStopBot).Platform(platform.Twitch).Build()
	_, err = conn.QueueSubscribe(topic, topic, c.StopBot)
	assert.NoError(err, "cannot subscribe to: "+topic)
}

func (c *BotController) StartBot(msg *nats.Msg) {
	var payload apptype.PlatformStartBot


	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.botService.StartBot(ctx, payload.Data)
	if err != nil {
		c.logger.DebugContext(ctx, "cannot start a bot", "payload", payload, "err", err)
		return
	}
}

func (c *BotController) StopBot(msg *nats.Msg) {
	var payload apptype.PlatformStopBot

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.botService.StopBot(ctx, payload.Data)
	if err != nil {
		c.logger.DebugContext(ctx, "cannot stop bot", "payload", payload, "err", err)
		return
	}
}
