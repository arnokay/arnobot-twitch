package service

import (
	"context"
	"log/slog"

	"arnobot-shared/data"
	"arnobot-shared/db"
	"arnobot-shared/pkg/errs"
	"arnobot-shared/storage"
)

type BotService struct {
	storage storage.Storager

	logger *slog.Logger
}

func NewBotService(store storage.Storager) *BotService {
	return &BotService{
		storage: store,
	}
}

func (s *BotService) DefaultBotGet(ctx context.Context) (*data.TwitchDefaultBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchDefaultBotGet(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "there is no default bot", "err", err)
		return nil, errs.ErrInternal
	}
	bot := data.NewTwitchDefaultBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) DefaultBotChange(ctx context.Context, twitchUserID string) error {
  count, err := s.storage.Query(ctx).TwitchDefaultBotUpdate(ctx, twitchUserID)
  if err != nil {
    s.logger.ErrorContext(ctx, "cannot update default bot", "err", err)
    return errs.ErrInternal
  }
  if count == 0 {
    s.logger.ErrorContext(ctx, "there is no default bot", "err", err)
    return errs.ErrInternal
  }

  return nil
}

func (s *BotService) SelectedBotGet(ctx context.Context, userID int32) (*data.TwitchSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotGet(ctx, userID)
	if err != nil {
		s.logger.DebugContext(ctx, "there is no selected bot", "err", err)
		return nil, errs.ErrNotFound
	}
	bot := data.NewTwitchSelectedBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) SelectedBotChange(ctx context.Context, bot data.TwitchBot) error {
  count, err := s.storage.Query(ctx).TwitchSelectedBotChange(ctx, db.TwitchSelectedBotChangeParams{
		UserID:       bot.UserID,
		TwitchUserID: bot.TwitchUserID,
	})
  if err != nil {
    s.logger.ErrorContext(ctx, "cannot change selected bot", "err", err)
    return errs.ErrInternal
  }
  if count == 0 {
    s.logger.ErrorContext(ctx, "there is no selected bot", "err", err)
    return errs.ErrInternal
  }

  return nil
}

