package service

import (
	"arnobot-shared/service"
)

type Services struct {
	AuthModuleService  *service.AuthModuleService
	CoreModuleService  *service.CoreModuleService
	HelixManager       *service.HelixManager
	BotService         *BotService
	WebhookService     *WebhookService
	TwitchService      *TwitchService
	TransactionService service.ITransactionService
}
