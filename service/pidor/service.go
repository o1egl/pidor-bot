//go:generate $TOOLS_BIN/mockgen -package mocks -source $GOFILE -destination ./mocks/$GOFILE
package pidor

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"

	"github.com/o1egl/pidor-bot/config"
	"github.com/o1egl/pidor-bot/log"
	"github.com/o1egl/pidor-bot/repo"
)

type Service struct {
	logger        log.Logger
	bot           TGBotAPI
	enableOpenAI  bool
	openAIClient  OpenAIClient
	repoClient    repo.Repo
	updates       tgbotapi.UpdatesChannel
	adminUsername string
	shutdownCh    chan struct{}
	doneCh        chan struct{}
}

type OpenAIClient interface {
	CreateChatCompletion(
		ctx context.Context,
		request openai.ChatCompletionRequest,
	) (openai.ChatCompletionResponse, error)
}

type TGBotAPI interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	StopReceivingUpdates()
}

func New(cfg *config.Config, logger log.Logger, repoClient repo.Repo) (*Service, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	bot.Debug = cfg.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = cfg.FetchingTimeout

	return &Service{
		logger:        logger,
		bot:           bot,
		enableOpenAI:  cfg.EnableOpenAI,
		openAIClient:  openai.NewClient(cfg.OpenAIToken),
		repoClient:    repoClient,
		updates:       bot.GetUpdatesChan(u),
		adminUsername: cfg.AdminUsername,
		shutdownCh:    make(chan struct{}),
		doneCh:        make(chan struct{}),
	}, nil
}

func (s *Service) Start() error {
	s.logger.Info("Pidor service started")
	defer func() {
		close(s.doneCh)
	}()

	ctx := context.Background()
	for update := range s.updates {
		logger := s.logger.With(zap.Int("update_id", update.UpdateID))
		ctx := log.ToContext(ctx, logger)
		go s.processUpdate(ctx, update)
	}
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping pidor service")
	defer s.logger.Info("Pidor service stopped")

	close(s.shutdownCh)
	s.bot.StopReceivingUpdates()

	select {
	case <-s.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Service) processUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	var err error
	switch {
	case strings.HasPrefix(update.Message.Text, "/regme"):
		err = s.handleReg(ctx, update)
	case strings.HasPrefix(update.Message.Text, "/pidor"):
		err = s.handlePidor(ctx, update)
	case strings.HasPrefix(update.Message.Text, "/stats_year"):
		err = s.handleStats(ctx, update, StatsPeriodYear)
	case strings.HasPrefix(update.Message.Text, "/stats_all"):
		err = s.handleStats(ctx, update, StatsPeriodAll)
	case strings.HasPrefix(update.Message.Text, "/stats"):
		err = s.handleStats(ctx, update, StatsPeriodMonth)
	case strings.HasPrefix(update.Message.Text, "/reset"):
		err = s.handleReset(ctx, update)
	default:
		err = s.handleUnknown(ctx, update)
	}

	if err != nil {
		log.FromContext(ctx).Error("Failed to process update", zap.Error(err))
		if err := s.sendMessage(update.Message.Chat.ID, fmt.Sprintf("Ощибка при обработке запроса: %d", update.UpdateID)); err != nil {
			log.FromContext(ctx).Error("Failed send message", zap.Error(err))
		}
	}
}
