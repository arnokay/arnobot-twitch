package service

import (
	"context"
	"log/slog"

	"arnobot-shared/applog"
	"arnobot-shared/data"
	"arnobot-shared/apperror"
	"arnobot-shared/service"
	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	helixManager *service.HelixManager
	logger       *slog.Logger
}

func NewTwitchService(
	helixManager *service.HelixManager,
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

func (s *TwitchService) GetBotChannelRole(
	ctx context.Context,
	botProvider data.AuthProvider,
	broadcasterID string,
) (data.TwitchBotRole, error) {
	client := s.helixManager.GetByProvider(ctx, botProvider)

	badges, err := client.GetChannelChatBadges(&helix.GetChatBadgeParams{
		BroadcasterID: broadcasterID,
	})
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"cannot request chat badges",
			"broadcasterID", broadcasterID,
			"botID", botProvider.ProviderUserID,
			"err", err,
		)
		return "", apperror.ErrExternal
	}

	role := data.TwitchBotRoleUser
	for _, badge := range badges.Data.Badges {
		if badge.SetID == "broadcaster" {
			role = data.TwitchBotRoleBroadcaster
			break
		}
		if badge.SetID == "moderator" {
			role = data.TwitchBotRoleModerator
			break
		}
		if badge.SetID == "vip" {
			role = data.TwitchBotRoleVIP
			break
		}
	}

	return role, nil
}
