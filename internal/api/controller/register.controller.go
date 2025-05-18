package controller

import (
	"log/slog"

	"arnobot-shared/applog"
	"github.com/labstack/echo/v4"
)

type RegisterController struct {
	logger *slog.Logger
  // middlewares *
}

func NewRegisterController() *RegisterController {
	logger := applog.NewServiceLogger("register-controller")

	return &RegisterController{
		logger: logger,
	}
}

func (c *RegisterController) Routes(parentGroup *echo.Group) {
	g := parentGroup.Group("/register")
  g.POST("/", c.Register)
}

func (c *RegisterController) Register(ctx echo.Context) error {
  return nil
}
