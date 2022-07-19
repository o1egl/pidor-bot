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

	if hasVotesWithin(votes, time.Hour) {
		return s.penaltyVote(ctx, update)
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

func (s *Service) penaltyVote(ctx context.Context, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	user := UserFromAPI(update.Message.From)

	if err := s.sendTyping(chatID); err != nil {
		return err
	}

	if err := s.repoClient.UpsertUser(ctx, chatID, user); err != nil {
		return err
	}

	if err := s.repoClient.CreateVote(ctx, chatID, domain.Vote{UserID: user.ID, Time: time.Now()}); err != nil {
		return err
	}

	time.Sleep(time.Second)

	if err := s.sendMessage(chatID, "Ах ты пидор! Нехуй было меня будить в неположенное время! Теперь пидором будет {{user}}!", NewMentionVar("{{user}}", user.Mention(), user.ID)); err != nil {
		return err
	}

	return nil
}

func hasVotesWithin(votes []domain.Vote, dur time.Duration) bool {
	latHour := time.Now().Truncate(dur)
	for _, vote := range votes {
		if vote.Time.Truncate(dur).Equal(latHour) {
			return true
		}
	}
	return false
}
