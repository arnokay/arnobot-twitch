package controller

import (
	"context"
	"time"

	"github.com/arnokay/arnobot-shared/controllers/mb"
	"github.com/arnokay/arnobot-shared/trace"
	"github.com/nats-io/nats.go"
)

type Controllers struct {
	ChatController controllers.NatsController
	BotController  controllers.NatsController
}

func (c *Controllers) Connect(conn *nats.Conn) {
	c.ChatController.Connect(conn)
	c.BotController.Connect(conn)
}

func newControllerContext(traceID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx = trace.Context(ctx, traceID)

	return ctx, cancel
}
