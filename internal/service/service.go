package service

import (
	"github.com/arnokay/arnobot-shared/service"
)

type Services struct {
	AuthModule         *service.AuthModule
	PlatformModule     *service.PlatformModuleOut
	HelixManager       *HelixManager
	BotService         *BotService
	WebhookService     *WebhookService
	TwitchService      *TwitchService
	TransactionService service.ITransactionService
}
