package controller

import (
	"github.com/labstack/echo/v4"
)

type Contollers struct {
	WebhookController *WebhookController
}

func (c *Contollers) Routes(parentGroup *echo.Group) {
	c.WebhookController.Routes(parentGroup)
}
