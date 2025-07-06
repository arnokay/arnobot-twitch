package service

import (
	"context"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/db"

	"github.com/nicklaw5/helix/v2"
)

type ChatService struct {
	helix *helix.Client
  querier db.Querier
  logger applog.Logger
}

type TwitchUserIDGetter interface {
  GetUserID() string
}

func (s *ChatService) GetDefaultBot(ctx context.Context) (*db.TwitchDefaultBot, error) {
  defaultBot, err := s.querier.TwitchDefaultBotGet(ctx)
  if err != nil {
    s.logger.Error("cannot get default bot", "err", err)
    return nil, apperror.ErrInternal
  }

  return &defaultBot, nil
}

func (s *ChatService) SubscribeChannelChatMessage(ctx context.Context, broadcaster TwitchUserIDGetter) {
  s.helix.CreateEventSubSubscription(&helix.EventSubSubscription{
    Type: "channel.chat.message",
    Version: "1",
    Condition: helix.EventSubCondition{
      BroadcasterUserID: broadcaster.GetUserID(),
      UserID: "",
    },
    Transport: helix.EventSubTransport{
      Method: "webhook",
      Callback: "tunnel.arnokay.com/",
      Secret: "secret-test",
    },
  })
}

func (s *ChatService) TwitchMsgToSystemMsg(msg helix.EventSubChannelChatMessageEvent) {
}
