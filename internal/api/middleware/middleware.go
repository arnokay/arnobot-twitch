package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/arnokay/arnobot-shared/applog"
	"github.com/arnokay/arnobot-shared/middlewares"
	"github.com/arnokay/arnobot-shared/apperror"

	"github.com/labstack/echo/v4"
	"github.com/nicklaw5/helix/v2"

	"github.com/arnokay/arnobot-twitch/internal/config"
)

type Middlewares struct {
	logger *slog.Logger

	AuthMiddlewares *middlewares.AuthMiddlewares
}

func New(
	authMiddlewares *middlewares.AuthMiddlewares,
) *Middlewares {
	logger := applog.NewServiceLogger("app-middleware")

	return &Middlewares{
		logger:          logger,
		AuthMiddlewares: authMiddlewares,
	}
}

func (m *Middlewares) VerifyTwitchWebhook(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: handle length of request body
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			m.logger.ErrorContext(c.Request().Context(), "cannot read body", "err", err)
			return apperror.ErrUnauthorized
		}
		c.Request().Body.Close()
		c.Request().Body = io.NopCloser(bytes.NewReader(body))

		// TODO: maybe move to db per webhook?
		if !helix.VerifyEventSubNotification(config.Config.Webhooks.Secret, c.Request().Header, string(body)) {
			m.logger.ErrorContext(c.Request().Context(), "unverified attempt to access webhook")
			return apperror.ErrUnauthorized
		}

		msgType := c.Request().Header.Get("Twitch-Eventsub-Message-Type")
		if msgType == "" {
			m.logger.ErrorContext(c.Request().Context(), "message type is empty")
			return apperror.ErrInvalidInput
		}

		if msgType != "webhook_callback_verification" {
			return next(c)
		}

		var event struct {
			Challenge    string                     `json:"challenge"`
			Subscription helix.EventSubSubscription `json:"subscription"`
		}
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&event)
		if err != nil {
			m.logger.ErrorContext(c.Request().Context(), "cannot decode body", "err", err)
			return apperror.ErrInvalidInput
		}

		if event.Challenge != "" {
      m.logger.DebugContext(c.Request().Context(), "confirmed challenge", "sub", event.Subscription.ID, "subType", event.Subscription.Type)
			return c.String(http.StatusOK, event.Challenge)
		}

		return next(c)
	}
}
