package controller

import (
	"github.com/labstack/echo/v4"
)

type Contollers struct {
	ChannelWebhookController *ChannelWebhookController
}

func (c *Contollers) Routes(parentGroup *echo.Group) {
	c.ChannelWebhookController.Routes(parentGroup)
}
