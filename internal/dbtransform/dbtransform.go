package dbtransform

import (
	"github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/db"
)

func NewPlatformDefaultBotFromDB(fromDB db.TwitchDefaultBot) data.PlatformDefaultBot {
	return data.PlatformDefaultBot{
		BotID: fromDB.BotID,
	}
}

func NewPlatformSelectedBotFromDB(fromDB db.TwitchSelectedBot) data.PlatformSelectedBot {
	return data.PlatformSelectedBot{
		UserID:        fromDB.UserID,
		BotID:         fromDB.BotID,
		BroadcasterID: fromDB.BroadcasterID,
		Enabled:       fromDB.Enabled,
		UpdatedAt:     fromDB.UpdatedAt,
	}
}

func NewPlatformBotFromDB(fromDB db.TwitchBot) data.PlatformBot {
	return data.PlatformBot{
		UserID:        fromDB.UserID,
		BroadcasterID: fromDB.BroadcasterID,
		BotID:         fromDB.BotID,
	}
}

func NewPlatformBotCreateToDB(d data.PlatformBotCreate) db.TwitchBotCreateParams {
	return db.TwitchBotCreateParams{
		UserID:        d.UserID,
		BotID:         d.BotID,
		BroadcasterID: d.BroadcasterID,
	}
}

func NewPlatformBotsGetToDB(d data.PlatformBotsGet) db.TwitchBotsGetParams {
	return db.TwitchBotsGetParams{
		UserID:        d.UserID,
		BotID:         d.BotID,
		BroadcasterID: d.BroadcasterID,
	}
}
