package service

import (
	"context"
	"log/slog"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	helixManager *HelixManager
	logger       *slog.Logger
}

func NewTwitchService(
	helixManager *HelixManager,
) *TwitchService {
	logger := applog.NewServiceLogger("twitch-service")

	return &TwitchService{
		helixManager: helixManager,
		logger:       logger,
	}
}

func (s *TwitchService) AppSendChannelMessage(
	ctx context.Context,
	botID string,
	broadcasterID string,
	message string,
	replyTo string,
) error {
	client := s.helixManager.GetApp(ctx)

	_, err := client.SendChatMessage(&helix.SendChatMessageParams{
		BroadcasterID:        broadcasterID,
		SenderID:             botID,
		Message:              message,
		ReplyParentMessageID: replyTo,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot send message to chat", "err", err, "broadcasterID", broadcasterID, "botID", botID, "message", message, "replyTo", replyTo)
		return apperror.ErrExternal
	}

	return nil
}
