package pidor

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) handleUnknown(ctx context.Context, update tgbotapi.Update) error {
	return nil
}
