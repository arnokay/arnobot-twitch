package service

import (
	"bytes"
	"context"
	
	"sync"
	"time"

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/pkg/assert"
	sharedService "github.com/arnokay/arnobot-shared/service"
	"github.com/arnokay/arnobot-shared/trace"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nicklaw5/helix/v2"
)

// TODO: right now there is no cleanup for clients
type HelixManager struct {
	logger       applog.Logger
	clientID     string
	clientSecret string

	appClient *helix.Client

	clients map[string]*helix.Client
	mu      sync.RWMutex

	authModule *sharedService.AuthModule
	cache      jetstream.KeyValue
}

func NewHelixManager(
	cache jetstream.KeyValue,
	authModule *sharedService.AuthModule,
	clientID, clientSecret string,
) *HelixManager {
	logger := applog.NewServiceLogger("helix-manager")

	appClient, err := helix.NewClient(&helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	})
	assert.NoError(err, "helix client needs to be initialized")

	token, err := appClient.RequestAppAccessToken([]string{
		"user:read:chat",
		"user:write:chat",
		"user:bot",
		"channel:bot",
		"bits:read",
	})
	assert.NoError(err, "cannot get access tokens for app client")
	appClient.SetAppAccessToken(token.Data.AccessToken)

	return &HelixManager{
		logger:       logger,
		clientID:     clientID,
		clientSecret: clientSecret,
		appClient:    appClient,
		clients:      make(map[string]*helix.Client),
		authModule:   authModule,
	}
}

func (hm *HelixManager) GetApp(ctx context.Context) *helix.Client {
	return hm.appClient
}

func (hm *HelixManager) GetByID(ctx context.Context, twitchID string) (*helix.Client, error) {
	hm.mu.RLock()
	client, exists := hm.clients[twitchID]
	hm.mu.RUnlock()

	if exists {
		return client, nil
	}

	return nil, apperror.New(apperror.CodeNotFound, "helix client is not found", nil)
}

func (hm *HelixManager) GetByProvider(ctx context.Context, provider data.AuthProvider) *helix.Client {
	hm.mu.RLock()
	client, exists := hm.clients[provider.ProviderUserID]
	hm.mu.RUnlock()

	if exists {
		return client
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if client, exists := hm.clients[provider.ProviderUserID]; exists {
		return client
	}

	client, _ = helix.NewClient(&helix.Options{
		ClientID:        hm.clientID,
		ClientSecret:    hm.clientSecret,
		UserAccessToken: provider.AccessToken,
		RefreshToken:    provider.RefreshToken,
	})

	client.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
    ctx = trace.Context(ctx, trace.New())
		// COMBAK: maybe set ttl?
		hm.cache.Put(
			ctx,
			"cm.art."+provider.Provider+"."+provider.ProviderUserID,
			bytes.Join([][]byte{[]byte(newAccessToken), []byte(newRefreshToken)}, []byte("...")),
		)
		hm.logger.InfoContext(ctx, "token refreshed", "providerUserID", provider.ProviderUserID)
		err := hm.authModule.AuthProviderUpdateTokens(ctx, data.AuthProviderUpdateTokens{
			ID:           provider.ID,
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		})
		if err != nil {
			hm.logger.ErrorContext(ctx, "failed to update tokens", "providerID", provider.ID, "providerUserID", provider.ProviderUserID)
		}
	})

	hm.clients[provider.ProviderUserID] = client

	return client
}
