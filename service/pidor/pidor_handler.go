package pidor

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/domain"
	"github.com/o1egl/pidor-bot/repo"
)

func (s *Service) handlePidor(ctx context.Context, update tgbotapi.Update) error {
	today := time.Now()
	votes, err := s.repoClient.GetVotes(
		ctx,
		update.Message.Chat.ID,
		repo.WithYear(today.Year()),
		repo.WithMonth(int(today.Month())),
		repo.WithDay(today.Day()),
	)
	if err != nil {
		return err
	}
	// Check if there is any vote for today
	if len(votes) > 0 {
		vote := votes[0]
		user, err := s.repoClient.GetUser(ctx, update.Message.Chat.ID, vote.UserID)
		if err != nil {
			return err
		}
		return s.sendMessage(
			update.Message.Chat.ID,
			"Сегодня пидором дня был выбран {{user}}",
			NewMentionVar("{{user}}", user.Mention(), user.ID),
		)
	}

	users, err := s.repoClient.GetUsers(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return s.sendMessage(update.Message.Chat.ID, "В этом чате нет пидоров")
	}

	i, err := cryptoRand(int64(len(users)))
	if err != nil {
		return err
	}

	user := users[i]
	err = s.repoClient.CreateVote(ctx, update.Message.Chat.ID, domain.Vote{
		UserID: user.ID,
		Time:   time.Now(),
	})
	if err != nil {
		return err
	}

	return s.sendMessage(
		update.Message.Chat.ID,
		"Поздравляю {{user}}, ты пидор!",
		NewMentionVar("{{user}}", user.Mention(), user.ID),
	)
}
