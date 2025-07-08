package controller

import (
	"context"
	"time"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/apptype"
	"github.com/arnokay/arnobot-shared/trace"
	"github.com/nats-io/nats.go"
)

type Controllers struct {
	ChatController *ChatController
	BotController  *BotController
}

func (c *Controllers) Connect(conn *nats.Conn) {
	c.ChatController.Connect(conn)
	c.BotController.Connect(conn)
}

func newControllerContext(traceID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	ctx = trace.Context(ctx, traceID)

	return ctx, cancel
}

func handleRequest[TReq, TResp any](
	msg *nats.Msg,
	handler func(context.Context, TReq) (TResp, error),
) {
	var payload apptype.Request[TReq]
	var response apptype.Response[TResp]

	err := payload.Decode(msg.Data)
	if err != nil {
		response.ToFailErr(apperror.New(apperror.CodeInternal, "cannot decode payload", err))
		b, _ := response.Encode()
		msg.Respond(b)
		return
	}
	response.TraceID = payload.TraceID

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	result, err := handler(ctx, payload.Data)
	if err != nil {
		response.ToFailErr(err)
	} else {
		response.ToSuccess(result)
	}

	b, _ := response.Encode()
	msg.Respond(b)
}

func handlePublish[TReq any](
	msg *nats.Msg,
	handler func(context.Context, TReq) error,
) {
	var payload apptype.Request[TReq]

	payload.Decode(msg.Data)

	ctx, cancel := newControllerContext(payload.TraceID)
	defer cancel()

	handler(ctx, payload.Data)
}
