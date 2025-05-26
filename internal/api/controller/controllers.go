package controller

import "github.com/labstack/echo/v4"

type Contollers struct {
	ChannelWebhookController *ChannelWebhookController
	RegisterController       *RegisterController
}

func (c *Contollers) Routes(parentGroup *echo.Group) {
	c.RegisterController.Routes(parentGroup)
	c.ChannelWebhookController
}
