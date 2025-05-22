package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"arnobot-shared/applog"
	"arnobot-shared/data"
	"arnobot-shared/pkg/errs"
	"arnobot-shared/service"
	"github.com/nicklaw5/helix/v2"

	"arnobot-twitch/internal/config"
)

func GetCallbackURL(event string) (*url.URL, error) {
	u := url.URL{
		Host:   config.Config.Global.BaseURL,
		Scheme: "https",
	}
	switch event {
	case helix.EventSubTypeChannelChatMessage:
		u.Path = "/twitch/callback/channel-chat-message"
	default:
		return nil, fmt.Errorf("cannot create url, unknown event: %s", event)
	}

	return &u, nil
}

type WebhookService struct {
	helixManager  *service.HelixManager
	twitchService *TwitchService

	logger *slog.Logger
}

func NewWebhookService(
	helixManager *service.HelixManager,
	twitchService *TwitchService,
) *WebhookService {
	logger := applog.NewServiceLogger("webhook-service")

	return &WebhookService{
		helixManager:  helixManager,
		twitchService: twitchService,
		logger:        logger,
	}
}

func (s *WebhookService) SubscribeChannelChat(ctx context.Context, botProvider data.AuthProvider, broadcasterID string) error {
	client := s.helixManager.GetByProvider(ctx, botProvider)

	callbackURL, err := GetCallbackURL(helix.EventSubTypeChannelChatMessage)
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot create callback url", "err", err)
		return errs.ErrInternal
	}

	_, err = client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelChatMessage,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
			UserID:            botProvider.ProviderUserID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: callbackURL.String(),
			Secret:   config.Config.Webhooks.Secret,
		},
	})
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"cannot create subscription",
			"event", helix.EventSubTypeChannelChatMessage,
			"botID", botProvider.ProviderUserID,
			"broadcasterID", broadcasterID,
		)
		return errs.ErrExternal
	}

	// TODO: should save subscription? also probably secret must be per subscription

	return nil
}

func (s *WebhookService) Subscribe(ctx context.Context, botProvider data.AuthProvider, broadcasterID string) error {
	role, err := s.twitchService.GetBotChannelRole(ctx, botProvider, broadcasterID)
	if err != nil {
		return err
	}

	if role == data.TwitchBotRoleBroadcaster {
		// TODO: implement
	}
	if role == data.TwitchBotRoleModerator {
		// TODO: implement
	}
	if role == data.TwitchBotRoleVIP {
		// TODO: implement
	}

	err = s.SubscribeChannelChat(ctx, botProvider, broadcasterID)
	if err != nil {
		return err
	}

	return nil
}
