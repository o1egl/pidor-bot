package pidor

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/domain"
	"github.com/o1egl/pidor-bot/repo"
)

func pidorPhrases() [][]string {
	return [][]string{
		{
			"БЛЯЯЯЯ! Хуле выдоебались до меня?",
			"Здорова заебал",
			"Чё надо блять?",
			"Поиск пидора запущен",
			"Опять? Лан, погнали",
			"Ха! Думаешь кого-то подставить?",
			"А не ты ли часом сам пидорок? проверим",
		},
		{
			"Лаадно, ща всё сделаю, так-так..",
			"Опа-опа, а кто это у нас тут?",
			"Начинаю поиски...",
			"Не, ну тут и думать даже не надо",
			"Бля, ну и так же ясно",
			"А вы чё, сами не знаете что ли кто это?",
		},
		{
			"Ну чё, {{user}}, тебя вычислили, пидорок! Готовь жопу",
			"Ну конечно же это {{user}}, кто-то удивлён?",
			"О, да это же {{user}}, известный в чате пидор!",
			"Бляяя, ну как так-то?? Мы думали что ты нормальны пацан, а ты пидор, {{user}}",
			"Ну что, поздравляю с честно заработанным званием главного пидора чата, {{user}}",
			"Ну ты и пидор, {{user}}",
			"Ебать ты пидор, {{user}}",
			"Попался, {{user}}, пидрила ебаная",
		},
	}
}

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
		return s.sendMessage(update.Message.Chat.ID, "Время выбора нового пидора еще не прошло")
	}

	users, err := s.repoClient.GetUsers(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return s.sendMessage(update.Message.Chat.ID, "В этом чате нет пидоров")
	}

	user, err := randUser(users)
	if err != nil {
		return err
	}

	err = s.repoClient.CreateVote(ctx, update.Message.Chat.ID, domain.Vote{
		UserID: user.ID,
		Time:   time.Now(),
	})
	if err != nil {
		return err
	}

	messages, err := getPidorMessages(user)
	if err != nil {
		return err
	}

	return s.sendMessages(update.Message.Chat.ID, messages, time.Second)
}

func getPidorMessages(user domain.User) ([]Message, error) {
	phrases := pidorPhrases()
	phrase0, err := randString(phrases[0])
	if err != nil {
		return nil, err
	}
	phrase1, err := randString(phrases[1])
	if err != nil {
		return nil, err
	}
	phrase2, err := randString(phrases[2])
	if err != nil {
		return nil, err
	}

	messages := []Message{
		{
			Text: phrase0,
		},
		{
			Text: phrase1,
		},
		{
			Text: phrase2,
			Vars: []MessageVar{NewMentionVar("{{user}}", user.Mention(), user.ID)},
		},
	}
	return messages, nil
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
