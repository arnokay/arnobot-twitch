package service

import (
	"arnobot-shared/service"
)

type Services struct {
	AuthModuleService  *service.AuthModuleService
	HelixManager       *service.HelixManager
	BotService         *BotService
	WebhookService     *WebhookService
	TwitchService      *TwitchService
	TransactionService service.ITransactionService
}
