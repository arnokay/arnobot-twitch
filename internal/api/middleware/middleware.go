package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"arnobot-shared/applog"
	"arnobot-shared/middlewares"
	"arnobot-shared/pkg/errs"
	"github.com/labstack/echo/v4"
	"github.com/nicklaw5/helix/v2"

	"arnobot-twitch/internal/config"
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
			return errs.ErrUnauthorized
		}
		c.Request().Body.Close()

		c.Request().Body = io.NopCloser(bytes.NewReader(body))

		// TODO: maybe move to db per webhook?
		if !helix.VerifyEventSubNotification(config.Config.Webhooks.Secret, c.Request().Header, string(body)) {
			m.logger.ErrorContext(c.Request().Context(), "unverified attempt to access webhook")
			return errs.ErrUnauthorized
		}

		msgType := c.Request().Header.Get("Twitch-Eventsub-Message-Type")
		if msgType == "" {
			m.logger.ErrorContext(c.Request().Context(), "message type is empty")
			return errs.ErrInvalidInput
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
			return errs.ErrInvalidInput
		}

		if event.Challenge != "" {
			return c.String(http.StatusOK, event.Challenge)
		}

		return next(c)
	}
}
