package controller

import (
	"log/slog"

	"arnobot-shared/applog"
	"github.com/labstack/echo/v4"

	"arnobot-twitch/internal/api/middleware"
)

type RegisterController struct {
	logger *slog.Logger

	middlewares *middleware.Middlewares
}

func NewRegisterController(middlewares *middleware.Middlewares) *RegisterController {
	logger := applog.NewServiceLogger("register-controller")

	return &RegisterController{
		logger:      logger,
		middlewares: middlewares,
	}
}

func (c *RegisterController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/register", c.middlewares.AuthMiddlewares.UserSessionGuard)
	g.POST("/", c.Register)
}

func (c *RegisterController) Register(ctx echo.Context) error {
	return nil
}
