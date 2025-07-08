package service

import (
	"context"
	"sync"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	sharedData "github.com/arnokay/arnobot-shared/data"
	"github.com/nicklaw5/helix/v2"

	"github.com/arnokay/arnobot-twitch/internal/config"
)

type WebhookService struct {
	helixManager  *HelixManager
	twitchService *TwitchService

	logger applog.Logger

	webhookToScopes map[string][]string
	callbackURL     string
}

func NewWebhookService(
	helixManager *HelixManager,
	twitchService *TwitchService,
) *WebhookService {
	logger := applog.NewServiceLogger("webhook-service")

	webhooksToScopes := map[string][]string{
		helix.EventSubTypeChannelChatMessage: {"user:read:chat"},
	}

	return &WebhookService{
		helixManager:    helixManager,
		twitchService:   twitchService,
		logger:          logger,
		webhookToScopes: webhooksToScopes,
		callbackURL:     config.Config.Webhooks.Callback,
	}
}

func (s *WebhookService) SubscribeChannelChatMessageBot(
	ctx context.Context,
	botID,
	broadcasterID string,
) error {
	client := s.helixManager.GetApp(ctx)

	// TODO: handle response, 4xx is not considered as error, probably should handle those in helixManager
	response, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeChannelChatMessage,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
			UserID:            botID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: s.callbackURL,
			Secret:   config.Config.Webhooks.Secret,
		},
	})

	if err != nil || response.StatusCode >= 400 {
		s.logger.ErrorContext(
			ctx,
			"cannot create subscription",
			"err", err,
			"err_msg", response.ErrorMessage,
			"event", helix.EventSubTypeChannelChatMessage,
			"botID", botID,
			"broadcasterID", broadcasterID,
		)
		return apperror.ErrExternal
	}

	return nil
}

func (s *WebhookService) SubscribeStreamOnline(
	ctx context.Context,
	botProvider sharedData.AuthProvider,
	broadcasterID string,
) error {
	event := helix.EventSubTypeStreamOnline

	client := s.helixManager.GetByProvider(ctx, botProvider)

	_, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    event,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: s.callbackURL,
			Secret:   config.Config.Webhooks.Secret,
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot subscribe to event", "err", err, "broadcaster_id", broadcasterID, "event", event)
		return apperror.ErrInternal
	}

	return nil
}

func (s *WebhookService) SubscribeStreamOffline(
	ctx context.Context,
	botProvider sharedData.AuthProvider,
	broadcasterID string,
) error {
	event := helix.EventSubTypeStreamOffline

	client := s.helixManager.GetByProvider(ctx, botProvider)

	_, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    event,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: s.callbackURL,
			Secret:   config.Config.Webhooks.Secret,
		},
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot subscribe to event", "err", err, "broadcaster_id", broadcasterID, "event", event)
		return apperror.ErrInternal
	}

	return nil
}

func (s *WebhookService) SubscribeChannelChatMessage(
	ctx context.Context,
	botProvider sharedData.AuthProvider,
	broadcasterID string,
) error {
	event := helix.EventSubTypeChannelChatMessage

	client := s.helixManager.GetByProvider(ctx, botProvider)

	_, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    event,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: broadcasterID,
			UserID:            botProvider.ProviderUserID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: s.callbackURL,
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
		return apperror.ErrExternal
	}

	return nil
}

func (s *WebhookService) Unsubscribe(
	ctx context.Context,
	subscriptionID string,
) error {
	client := s.helixManager.GetApp(ctx)

	response, err := client.RemoveEventSubSubscription(subscriptionID)
	if err != nil || response.StatusCode >= 400 {
		s.logger.ErrorContext(
			ctx,
			"cannot unsubscribe",
			"err", err,
			"err_msg", response.ErrorMessage,
			"subscription_id", subscriptionID,
		)
		return apperror.ErrExternal
	}

	return nil
}

func (s *WebhookService) UnsubscribeAllBot(
	ctx context.Context,
	botID,
	broadcasterID string,
) error {
	client := s.helixManager.GetApp(ctx)

	var subIds []string

	var cursor string
	for {
		subs, err := client.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{
			UserID: botID,
			After:  cursor,
		})
		if err != nil || subs.StatusCode >= 400 {
			s.logger.ErrorContext(ctx, "error getting eventsub subscriptions", "err", err, "err_msg", subs.ErrorMessage)
			return apperror.ErrExternal
		}

		if len(subs.Data.EventSubSubscriptions) == 0 {
			s.logger.DebugContext(ctx, "no more eventsub subscriptions")
			break
		}

		for _, sub := range subs.Data.EventSubSubscriptions {
			if sub.Condition.BroadcasterUserID == broadcasterID {
				subIds = append(subIds, sub.ID)
			}
		}

		if subs.Data.Pagination.Cursor == "" {
			s.logger.DebugContext(ctx, "no more eventsub subscriptions")
			break
		}

		cursor = subs.Data.Pagination.Cursor
	}

	wg := sync.WaitGroup{}
	wg.Add(len(subIds))
	for _, id := range subIds {
		go func() {
			err := s.Unsubscribe(ctx, id)
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to unsubscribe", "err", err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func (s *WebhookService) Subscribe(
	ctx context.Context,
	botProvider sharedData.AuthProvider,
	broadcasterID string,
) error {
  var errs []error
	err := s.SubscribeChannelChatMessage(ctx, botProvider, broadcasterID)
	if err != nil {
		errs = append(errs, err)
	}
  err = s.SubscribeStreamOnline(ctx, botProvider, broadcasterID)
  if err != nil {
    errs = append(errs, err)
  }
  err = s.SubscribeStreamOffline(ctx, botProvider, broadcasterID)
  if err != nil {
    errs = append(errs, err)
  }

  if len(errs) != 0 {
    s.logger.ErrorContext(ctx, "cannot subscribe to all of them", "missed_subscriptions", len(errs))
    return apperror.ErrExternal
  }

	return nil
}
