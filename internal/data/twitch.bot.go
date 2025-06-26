package data

import (
	"time"

	"github.com/google/uuid"

	"github.com/arnokay/arnobot-shared/db"
)

type PlatformDefaultBot struct {
	BotID string
}

func NewPlatformDefaultBotFromDB(fromDB db.TwitchDefaultBot) PlatformDefaultBot {
	return PlatformDefaultBot{
		BotID: fromDB.BotID,
	}
}

type PlatformSelectedBot struct {
	UserID        uuid.UUID
	BotID         string
	BroadcasterID string
	UpdatedAt     time.Time
}

func NewPlatformSelectedBotFromDB(fromDB db.TwitchSelectedBot) PlatformSelectedBot {
  return PlatformSelectedBot{
    UserID: fromDB.UserID,
    BotID: fromDB.BotID,
    BroadcasterID: fromDB.BroadcasterID,
    UpdatedAt: fromDB.UpdatedAt,
  }
}

type PlatformBot struct {
	UserID        uuid.UUID
	BotID         string
	BroadcasterID string
}

func NewPlatformBotFromDB(fromDB db.TwitchBot) PlatformBot {
	return PlatformBot{
		UserID:        fromDB.UserID,
		BroadcasterID: fromDB.BroadcasterID,
		BotID:         fromDB.BotID,
	}
}

type PlatformBotCreate struct {
	UserID        uuid.UUID
	BotID         string
	BroadcasterID string
}

func (d PlatformBotCreate) ToDB() db.TwitchBotCreateParams {
	return db.TwitchBotCreateParams{
		UserID:        d.UserID,
		BotID:         d.BotID,
		BroadcasterID: d.BroadcasterID,
	}
}

type PlatformBotsGet struct {
	UserID        *uuid.UUID
	BotID         *string
	BroadcasterID *string
}

func (d PlatformBotsGet) ToDB() db.TwitchBotsGetParams {
	return db.TwitchBotsGetParams{
		UserID:        d.UserID,
		BotID:         d.BotID,
		BroadcasterID: d.BroadcasterID,
	}
}
