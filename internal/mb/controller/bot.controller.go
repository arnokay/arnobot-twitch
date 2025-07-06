package controller

import (
	

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/apptype"
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	"github.com/arnokay/arnobot-shared/platform"
	"github.com/arnokay/arnobot-shared/topics"
	"github.com/nats-io/nats.go"

	"github.com/arnokay/arnobot-twitch/internal/service"
)

type BotController struct {
	botService *service.BotService

	logger applog.Logger
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
	topic = topics.TopicBuilder(topics.PlatformGetBot).Platform(platform.Twitch).Build()
	_, err = conn.QueueSubscribe(topic, topic, c.GetBot)
	assert.NoError(err, "cannot subscribe to: "+topic)
}

func (c *BotController) GetBot(msg *nats.Msg) {
	var payload apptype.Request[data.PlatformBotGet]
	var response apptype.Response[data.PlatformBot]

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	bot, err := c.botService.SelectedBotGet(ctx, payload.Data.UserID)
	if err != nil {
		c.logger.DebugContext(ctx, "cannot get selected bot", "err", err)
		response.ToFailErr(err)
		b, _ := response.Encode()
		err := msg.Respond(b)
		c.logger.DebugContext(ctx, "cannot respond", "err", err)
		return
	}

	response.ToSuccess(data.PlatformBot{
		Platform: platform.Twitch,
		BotID:    bot.BotID,
		UserID:   bot.UserID,
		Enabled:  bot.Enabled,
	})
	b, _ := response.Encode()
	msg.Respond(b)
}

func (c *BotController) StartBot(msg *nats.Msg) {
	var payload apptype.Request[data.PlatformBotToggle]
	var response apptype.EmptyResponse

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.botService.StartBot(ctx, payload.Data)
	if err != nil {
		c.logger.DebugContext(ctx, "cannot start a bot", "payload", payload, "err", err)
		response.ToFailErr(err)
		b, _ := response.Encode()
		msg.Respond(b)
		return
	}

	response.ToSuccess(true)
	b, _ := response.Encode()
	msg.Respond(b)
}

func (c *BotController) StopBot(msg *nats.Msg) {
	var payload apptype.Request[data.PlatformBotToggle]
	var response apptype.EmptyResponse

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	err := c.botService.StopBot(ctx, payload.Data)
	if err != nil {
		c.logger.DebugContext(ctx, "cannot stop a bot", "payload", payload, "err", err)
		response.ToFailErr(err)
		b, _ := response.Encode()
		msg.Respond(b)
		return
	}

	response.ToSuccess(true)
	b, _ := response.Encode()
	msg.Respond(b)
}
