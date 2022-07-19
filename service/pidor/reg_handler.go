package pidor

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/domain"
)

func (s *Service) handleReg(ctx context.Context, update tgbotapi.Update) error {
	user := UserFromAPI(update.Message.From)
	err := s.repoClient.UpsertUser(ctx, update.Message.Chat.ID, user)
	if err != nil {
		return err
	}

	return s.sendMessage(
		update.Message.Chat.ID,
		"Поздравляю {{user}}, ты зарегистрировался в почетные ряды пидоров!",
		NewMentionVar("{{user}}", user.Mention(), user.ID),
	)
}

func UserFromAPI(user *tgbotapi.User) domain.User {
	return domain.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.UserName,
		IsActive:  true,
	}
}
