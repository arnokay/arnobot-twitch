package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"

	"arnobot-shared/applog"
	"arnobot-shared/data"
	"arnobot-shared/pkg/errs"
	"arnobot-shared/service"
	"github.com/nicklaw5/helix/v2"

	"arnobot-twitch/internal/config"
)

type WebhookService struct {
	helixManager  *service.HelixManager
	twitchService *TwitchService

	logger *slog.Logger

	webhookToScopes map[string][]string
	eventToCallback map[string]string
}

func NewWebhookService(
	helixManager *service.HelixManager,
	twitchService *TwitchService,
) *WebhookService {
	logger := applog.NewServiceLogger("webhook-service")

	webhooksToScopes := map[string][]string{
		helix.EventSubTypeChannelChatMessage: {"user:read:chat"},
	}

	eventToCallback := map[string]string{
		helix.EventSubTypeChannelChatMessage: "/twitch/callback/channel-chat-message",
	}

	return &WebhookService{
		helixManager:    helixManager,
		twitchService:   twitchService,
		logger:          logger,
		webhookToScopes: webhooksToScopes,
		eventToCallback: eventToCallback,
	}
}

func (s *WebhookService) canSubscribe(
	ctx context.Context,
	botProvider data.AuthProvider,
	event string,
) error {
	requiredScopes, ok := s.webhookToScopes[event]
	if !ok {
		return nil
	}

	missingScopes := make([]string, len(requiredScopes))

	for _, scope := range requiredScopes {
		if !slices.Contains(botProvider.Scopes, scope) {
			missingScopes = append(missingScopes, scope)
		}
	}

	if len(missingScopes) != 0 {
		s.logger.ErrorContext(
			ctx,
			"cannot subscribe to channel.chat.message, missing scope",
			"botID", botProvider.ProviderUserID,
			"botScopes", botProvider.Scopes,
			"webhookScopes", requiredScopes,
			"missingScopes", missingScopes,
		)
		return errs.New(
			errs.CodeForbidden,
			fmt.Sprintf("cannot subscribe user to %s because user missing scopes: %v", event, missingScopes),
			nil,
		)
	}

	return nil
}

func (s *WebhookService) SubscribeChannelChatMessage(
	ctx context.Context,
	botProvider data.AuthProvider,
	broadcasterID string,
) error {
	event := helix.EventSubTypeChannelChatMessage

	if err := s.canSubscribe(ctx, botProvider, event); err != nil {
		return err
	}

	client := s.helixManager.GetByProvider(ctx, botProvider)

	callbackURL, err := s.getCallbackURL(event)
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
			"err", err,
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

	err = s.SubscribeChannelChatMessage(ctx, botProvider, broadcasterID)
	if err != nil {
		return err
	}

	return nil
}

func (s *WebhookService) getCallbackURL(event string) (*url.URL, error) {
	u := url.URL{
		Host:   config.Config.Global.BaseURL,
		Scheme: "https",
	}
	path, ok := s.eventToCallback[event]
	if !ok {
		return nil, fmt.Errorf("no callback for specified event: %s", event)
	}

	u.Path = path

	return &u, nil
}
