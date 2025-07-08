package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/nicklaw5/helix/v2"

	"github.com/arnokay/arnobot-twitch/internal/config"
)

type EventSubscriptionRequest struct {
	EventType     string
	BroadcasterID string
	UserID        string
	Version       string
}

type SubscriptionResult struct {
	EventType string
	Error     error
}

type WebhookService struct {
	helixManager  *HelixManager
	twitchService *TwitchService
	logger        applog.Logger
	callbackURL   string
	secret        string
}

func NewWebhookService(
	helixManager *HelixManager,
	twitchService *TwitchService,
) *WebhookService {
	logger := applog.NewServiceLogger("webhook-service")

	return &WebhookService{
		helixManager:  helixManager,
		twitchService: twitchService,
		logger:        logger,
		callbackURL:   config.Config.Webhooks.Callback,
		secret:        config.Config.Webhooks.Secret,
	}
}

func (s *WebhookService) createEventSubscription(
	ctx context.Context,
	client *helix.Client,
	req EventSubscriptionRequest,
) error {
	condition := helix.EventSubCondition{
		BroadcasterUserID: req.BroadcasterID,
		UserID:            req.UserID,
	}

  if req.Version == "" {
    req.Version = "1"
  }

	subscription := &helix.EventSubSubscription{
		Type:      req.EventType,
		Version:   req.Version,
		Condition: condition,
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: s.callbackURL,
			Secret:   s.secret,
		},
	}

	response, err := client.CreateEventSubSubscription(subscription)
	if err != nil {
		return apperror.New(apperror.CodeExternal, "failed to create event subscription", err)
	}

	if response.StatusCode >= 400 {
		return apperror.New(apperror.CodeExternal, fmt.Sprintf("subscription failed with status %d: %s", response.StatusCode, response.ErrorMessage), nil)
	}

	return nil
}

func (s *WebhookService) SubscribeChannelChatMessage(ctx context.Context, botID, broadcasterID string) error {
	client := s.helixManager.GetApp(ctx)

	err := s.createEventSubscription(ctx, client, EventSubscriptionRequest{
		EventType:     helix.EventSubTypeChannelChatMessage,
		BroadcasterID: broadcasterID,
		UserID:        botID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot create chat message subscription",
			"err", err,
			"botID", botID,
			"broadcasterID", broadcasterID,
		)
		return apperror.New(apperror.CodeExternal, "failed to subscribe to chat messages", err)
	}

	return nil
}

func (s *WebhookService) SubscribeStreamOnline(ctx context.Context, broadcasterID string) error {
	client := s.helixManager.GetApp(ctx)

	err := s.createEventSubscription(ctx, client, EventSubscriptionRequest{
		EventType:     helix.EventSubTypeStreamOnline,
		BroadcasterID: broadcasterID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot create stream online subscription",
			"err", err,
			"broadcasterID", broadcasterID,
		)
		return apperror.New(apperror.CodeExternal, "failed to subscribe to stream online events", err)
	}

	return nil
}

func (s *WebhookService) SubscribeStreamOffline(ctx context.Context, broadcasterID string) error {
	client := s.helixManager.GetApp(ctx)

	err := s.createEventSubscription(ctx, client, EventSubscriptionRequest{
		EventType:     helix.EventSubTypeStreamOffline,
		BroadcasterID: broadcasterID,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot create stream offline subscription",
			"err", err,
			"broadcasterID", broadcasterID,
		)
		return apperror.New(apperror.CodeExternal, "failed to subscribe to stream offline events", err)
	}

	return nil
}

func (s *WebhookService) Unsubscribe(ctx context.Context, subscriptionID string) error {
	client := s.helixManager.GetApp(ctx)

	response, err := client.RemoveEventSubSubscription(subscriptionID)
	if err != nil {
		return apperror.New(apperror.CodeExternal, "failed to remove event subscription", err)
	}

	if response.StatusCode >= 400 {
		s.logger.ErrorContext(ctx, "cannot unsubscribe",
			"err", err,
			"err_msg", response.ErrorMessage,
			"subscription_id", subscriptionID,
		)
		return apperror.New(apperror.CodeExternal, fmt.Sprintf("unsubscribe failed with status %d: %s", response.StatusCode, response.ErrorMessage), nil)
	}

	return nil
}

func (s *WebhookService) UnsubscribeAllBot(ctx context.Context, botID, broadcasterID string) error {
	client := s.helixManager.GetApp(ctx)

	subscriptionIDs, err := s.getSubscriptionIDs(ctx, client, botID, broadcasterID)
	if err != nil {
		return apperror.New(apperror.CodeExternal, "failed to get subscription IDs", err)
	}

	if len(subscriptionIDs) == 0 {
		s.logger.DebugContext(ctx, "no subscriptions to unsubscribe")
		return nil
	}

	return s.unsubscribeAll(ctx, subscriptionIDs)
}

func (s *WebhookService) getSubscriptionIDs(ctx context.Context, client *helix.Client, botID, broadcasterID string) ([]string, error) {
	var subscriptionIDs []string
	var cursor string

	for {
		subs, err := client.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{
			UserID: botID,
			After:  cursor,
		})
		if err != nil {
			return nil, apperror.New(apperror.CodeExternal, "failed to get event subscriptions", err)
		}

		if subs.StatusCode >= 400 {
			return nil, apperror.New(apperror.CodeExternal, fmt.Sprintf("failed to get subscriptions with status %d: %s", subs.StatusCode, subs.ErrorMessage), nil)
		}

		if len(subs.Data.EventSubSubscriptions) == 0 {
			break
		}

		for _, sub := range subs.Data.EventSubSubscriptions {
			if sub.Condition.BroadcasterUserID == broadcasterID {
				subscriptionIDs = append(subscriptionIDs, sub.ID)
			}
		}

		if subs.Data.Pagination.Cursor == "" {
			break
		}

		cursor = subs.Data.Pagination.Cursor
	}

	return subscriptionIDs, nil
}

func (s *WebhookService) unsubscribeAll(ctx context.Context, subscriptionIDs []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(subscriptionIDs))

	for _, id := range subscriptionIDs {
		wg.Add(1)
		go func(subscriptionID string) {
			defer wg.Done()
			if err := s.Unsubscribe(ctx, subscriptionID); err != nil {
				errChan <- apperror.New(apperror.CodeExternal, fmt.Sprintf("failed to unsubscribe %s", subscriptionID), err)
			}
		}(id)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		s.logger.ErrorContext(ctx, "some unsubscriptions failed", "failed_count", len(errs))
		return apperror.New(apperror.CodeExternal, "some unsubscriptions failed", errors.Join(errs...))
	}

	return nil
}

func (s *WebhookService) SubscribeAll(ctx context.Context, botID string, broadcasterID string) error {
	subscriptions := []struct {
		name string
		fn   func() error
	}{
		{"chat_message", func() error { return s.SubscribeChannelChatMessage(ctx, botID, broadcasterID) }},
		{"stream_online", func() error { return s.SubscribeStreamOnline(ctx, broadcasterID) }},
		{"stream_offline", func() error { return s.SubscribeStreamOffline(ctx, broadcasterID) }},
	}

	var results []SubscriptionResult
	for _, sub := range subscriptions {
		err := sub.fn()
		results = append(results, SubscriptionResult{
			EventType: sub.name,
			Error:     err,
		})
	}

	var failedSubs []string
	for _, result := range results {
		if result.Error != nil {
			failedSubs = append(failedSubs, result.EventType)
		}
	}

	if len(failedSubs) > 0 {
		s.logger.ErrorContext(ctx, "failed to subscribe to some events",
			"failed_subscriptions", failedSubs,
			"total_failed", len(failedSubs),
		)
		return apperror.New(apperror.CodeExternal, "failed to subscribe to some events", nil)
	}

	s.logger.InfoContext(ctx, "successfully subscribed to all events",
		"botID", botID,
		"broadcasterID", broadcasterID,
	)

	return nil
}
