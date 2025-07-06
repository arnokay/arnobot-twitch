package service

import (
	"context"
	

	"github.com/arnokay/arnobot-shared/apperror"
	"github.com/arnokay/arnobot-shared/applog"
	sharedData "github.com/arnokay/arnobot-shared/data"
	"github.com/arnokay/arnobot-shared/db"
	sharedService "github.com/arnokay/arnobot-shared/service"
	"github.com/arnokay/arnobot-shared/storage"
	"github.com/google/uuid"

	"github.com/arnokay/arnobot-twitch/internal/data"
)

type BotService struct {
	storage       storage.Storager
	txService     sharedService.ITransactionService
	authModule    *sharedService.AuthModule
	whService     *WebhookService
	twitchService *TwitchService

	logger applog.Logger
}

func NewBotService(
	store storage.Storager,
	txService sharedService.ITransactionService,
	authModule *sharedService.AuthModule,
	whService *WebhookService,
	twitchService *TwitchService,
) *BotService {
	logger := applog.NewServiceLogger("bot-service")
	return &BotService{
		storage:       store,
		txService:     txService,
		authModule:    authModule,
		whService:     whService,
		twitchService: twitchService,
		logger:        logger,
	}
}

func (s *BotService) SelectedBotChangeStatus(ctx context.Context, userID uuid.UUID, enable bool) error {
	_, err := s.storage.Query(ctx).TwitchSelectedBotStatusChange(ctx, db.TwitchSelectedBotStatusChangeParams{
		UserID:  userID,
		Enabled: enable,
	})
	if err != nil {
		s.logger.DebugContext(ctx, "cannot enable/disable bot", "err", err)
		return s.storage.HandleErr(ctx, err)
	}

	return nil
}


func (s *BotService) StartBot(ctx context.Context, arg sharedData.PlatformBotToggle) error {
	txCtx, err := s.txService.Begin(ctx)
	defer s.txService.Rollback(txCtx)
	if err != nil {
		return err
	}

	selectedBot, err := s.SelectedBotGet(txCtx, arg.UserID)
	if err != nil {
		if !apperror.IsAppErr(err) {
			return err
		}

		selectedBot, err = s.SelectedBotSetDefault(txCtx, arg.UserID)
		if err != nil {
			return err
		}
	}

	err = s.txService.Commit(txCtx)
	if err != nil {
		return err
	}

	err = s.whService.SubscribeChannelChatMessageBot(txCtx, selectedBot.BotID, selectedBot.BroadcasterID)
	if err != nil {
		return err
	}
	err = s.SelectedBotChangeStatus(ctx, arg.UserID, true)
	if err != nil {
		return err
	}

	s.twitchService.AppSendChannelMessage(ctx, selectedBot.BotID, selectedBot.BroadcasterID, "hi!", "")

	return nil
}

func (s *BotService) StopBot(ctx context.Context, arg sharedData.PlatformBotToggle) error {
	selectedBot, err := s.SelectedBotGet(ctx, arg.UserID)
	if err != nil {
		return err
	}

	err = s.whService.UnsubscribeAllBot(ctx, selectedBot.BotID, selectedBot.BroadcasterID)
	if err != nil {
		s.logger.DebugContext(ctx, "bot cannot unsubscribe")
		return err
	}

	return nil
}

func (s *BotService) SelectedBotSetDefault(ctx context.Context, userID uuid.UUID) (*data.PlatformSelectedBot, error) {
	var bot *data.PlatformBot

	txCtx, err := s.txService.Begin(ctx)
	defer s.txService.Rollback(txCtx)
	if err != nil {
		return nil, err
	}

	bots, err := s.BotsGet(ctx, data.PlatformBotsGet{
		UserID: &userID,
	})
	if err != nil {
		return nil, err
	}

	if len(bots) != 0 {
		bot = &bots[0]
	} else {
		defaultBot, err := s.DefaultBotGet(ctx)
		if err != nil {
			return nil, err
		}
		userProvider, err := s.authModule.AuthProviderGet(ctx, sharedData.AuthProviderGet{
			UserID:   &userID,
			Provider: "twitch",
		})
		if err != nil {
			return nil, err
		}
		bot, err = s.BotCreate(ctx, data.PlatformBotCreate{
			UserID:        userID,
			BotID:         defaultBot.BotID,
			BroadcasterID: userProvider.ProviderUserID,
		})
		if err != nil {
			return nil, err
		}
	}

	selectedBot, err := s.SelectedBotChange(ctx, *bot)
	if err != nil {
		return nil, err
	}

	err = s.txService.Commit(txCtx)
	if err != nil {
		return nil, err
	}

	return selectedBot, nil
}

func (s *BotService) SelectedBotGetByBroadcasterID(ctx context.Context, broadcasterID string) (*data.PlatformSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotGetByBroadcasterID(ctx, broadcasterID)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot get selected bot", "err", err, "broadcasterID", broadcasterID)
		return nil, s.storage.HandleErr(ctx, err)
	}

	bot := data.NewPlatformSelectedBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) BotCreate(ctx context.Context, arg data.PlatformBotCreate) (*data.PlatformBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchBotCreate(ctx, arg.ToDB())
	if err != nil {
		s.logger.DebugContext(ctx, "cannot create bot", "err", err)
		return nil, s.storage.HandleErr(ctx, err)
	}

	bot := data.NewPlatformBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) BotsGet(ctx context.Context, arg data.PlatformBotsGet) ([]data.PlatformBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchBotsGet(ctx, arg.ToDB())
	if err != nil {
		s.logger.ErrorContext(ctx, "cannot get twitch bots")
		return nil, s.storage.HandleErr(ctx, err)
	}

	var bots []data.PlatformBot
	for _, bot := range fromDB {
		bots = append(bots, data.NewPlatformBotFromDB(bot))
	}

	return bots, nil
}

func (s *BotService) DefaultBotGet(ctx context.Context) (*data.PlatformDefaultBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchDefaultBotGet(ctx)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot get default bot")
		return nil, s.storage.HandleErr(ctx, err)
	}
	bot := data.NewPlatformDefaultBotFromDB(fromDB)

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

func (s *BotService) SelectedBotGet(ctx context.Context, userID uuid.UUID) (*data.PlatformSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotGetByUserID(ctx, userID)
	if err != nil {
		s.logger.DebugContext(ctx, "cannot get selected bot")
		return nil, s.storage.HandleErr(ctx, err)
	}
	bot := data.NewPlatformSelectedBotFromDB(fromDB)

	return &bot, nil
}

func (s *BotService) SelectedBotChange(ctx context.Context, bot data.PlatformBot) (*data.PlatformSelectedBot, error) {
	fromDB, err := s.storage.Query(ctx).TwitchSelectedBotChange(ctx, db.TwitchSelectedBotChangeParams{
		UserID:        bot.UserID,
		BotID:         bot.BotID,
		BroadcasterID: bot.BroadcasterID,
	})
	if err != nil {
		s.logger.DebugContext(ctx, "cannot change selected bot", "err", err)
		return nil, s.storage.HandleErr(ctx, err)
	}

	selectedBot := data.NewPlatformSelectedBotFromDB(fromDB)

	return &selectedBot, nil
}
