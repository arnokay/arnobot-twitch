package service

import (
	"context"
	"log/slog"

	"arnobot-shared/applog"
	"arnobot-shared/data"
	"arnobot-shared/db"
	"arnobot-shared/apperror"
	"arnobot-shared/storage"

	"github.com/google/uuid"
)

type BotService struct {
	storage storage.Storager

	logger *slog.Logger
}

func NewBotService(store storage.Storager) *BotService {
	logger := applog.NewServiceLogger("bot-service")
	return &BotService{
		storage: store,
		logger:  logger,
	}
}

func (s *BotService) SelectedBotGetByBroadcasterID(ctx context.Context, broadcasterID string) (*data.TwitchSelectedBot, error) {
  fromDB, err := s.storage.Query(ctx).TwitchSelectedBotGetByBroadcasterID(ctx, broadcasterID)
  if err != nil {
    s.logger.DebugContext(ctx, "cannot get selected bot", "err", err, "broadcasterID", broadcasterID)
    return nil, s.storage.HandleErr(ctx, err)
  }

  bot := data.NewTwitchSelectedBotFromDB(fromDB)

  return &bot, nil
}

func (s *BotService) BotCreate(ctx context.Context, arg data.TwitchBotCreate) (*data.TwitchBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchBotCreate(ctx, arg.ToDB())
	if err != nil {
		s.logger.DebugContext(ctx, "cannot create bot", "err", err)
		return nil, s.storage.HandleErr(ctx, err)
	}

	bot := data.NewTwitchBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) BotsGet(ctx context.Context, arg data.TwitchBotsGet) ([]data.TwitchBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchBotsGet(ctx, arg.ToDB())
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot get twitch bots")
		return nil, s.storage.HandleErr(ctx, err)
	}

	var bots []data.TwitchBot
	for _, bot := range fromDB {
		bots = append(bots, data.NewTwitchBotFromDB(bot))
	}

	return bots, nil
}

func (s *BotService) DefaultBotGet(ctx context.Context) (*data.TwitchDefaultBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchDefaultBotGet(ctx)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot get default bot")
		return nil, s.storage.HandleErr(ctx, err)
	}
	bot := data.NewTwitchDefaultBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) DefaultBotChange(ctx context.Context, botID string) error {
	count, err := s.storage.Query(ctx).TwitchDefaultBotUpdate(ctx, botID)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot update default bot", "err", err)
		return s.storage.HandleErr(ctx, err)
	}

	if count == 0 {
		s.logger.ErrorContext(ctx, "there is no default bot to update???")
		return apperror.ErrInternal
	}

	return nil
}

func (s *BotService) SelectedBotGet(ctx context.Context, userID uuid.UUID) (*data.TwitchSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotGetByUserID(ctx, userID)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot get selected bot")
		return nil, s.storage.HandleErr(ctx, err)
	}
	bot := data.NewTwitchSelectedBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) SelectedBotChange(ctx context.Context, bot data.TwitchBot) (*data.TwitchSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotChange(ctx, db.TwitchSelectedBotChangeParams{
		UserID:        bot.UserID,
		BotID:         bot.BotID,
		BroadcasterID: bot.BroadcasterID,
	})
	if err != nil {
		s.logger.DebugContext(ctx, "cannot change selected bot", "err", err)
		return nil, s.storage.HandleErr(ctx, err)
	}

	selectedBot := data.NewTwitchSelectedBotFromDB(fromDB)

	return &selectedBot, nil
}
