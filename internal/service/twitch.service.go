package service

import (
	"context"

	"arnobot-shared/data"
	"arnobot-shared/service"
	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	helixManager *service.HelixManager
}

func NewTwitchService(
	helixManager *service.HelixManager,
) *TwitchService {
	return &TwitchService{
		helixManager: helixManager,
	}
}

func (s *WebhookService) GetChannelRole(
	ctx context.Context,
	botProvider data.AuthProvider,
	broadcasterID string,
) (data.TwitchBotRole, error) {
	client := s.helixManager.GetByProvider(ctx, botProvider)

	badges, err := client.GetChannelChatBadges(&helix.GetChatBadgeParams{
		BroadcasterID: broadcasterID,
	})
	if err != nil {
		return "", err
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
