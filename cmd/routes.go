package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *application) SetRoutes(e *echo.Group) {
	e.POST("/twitch/channel/chat/message", a.controllers.ChatController.ChannelChatMessageHandler)

	e.GET("/healthcheck", func(c echo.Context) error {
		response := make(map[string]string, 4)
		responseStatus := http.StatusOK

		response["messageBroker"] = "OK"
		if !a.msgBroker.IsConnected() {
			response["messageBroker"] = "FAIL"
		}

		for _, value := range response {
			if value != "OK" {
				responseStatus = http.StatusServiceUnavailable
			}

		}

		return c.JSON(responseStatus, response)
	})
}
