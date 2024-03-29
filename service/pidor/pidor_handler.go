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
			"БЛЯЯЯЯ! Хуле вы доебались до меня?",
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
			`
.∧＿∧ 
( ･ω･｡)つ━☆・*。 
⊂　 ノ 　　　・゜+. 
しーＪ　　　°。+ *´¨) 
　　　　　　　　　.· ´¸.·*´¨) 
　　　　　　　　　　(¸.·´ (¸.·'* ☆ ВЖУХ И ТЫ ПИДОР, {{user}}`,
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

	votingUserID := update.Message.From.ID

	if containsVotesFromUser(votes, votingUserID) {
		return s.sendMessage(update.Message.Chat.ID, "Пошел нахуй, пидорок, ты уже голосовал сегодня")
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
		UserID:      user.ID,
		VotedUserID: votingUserID,
		Time:        time.Now(),
	})
	if err != nil {
		return err
	}

	if err := s.sendTyping(update.Message.Chat.ID); err != nil {
		return err
	}

	messages, err := s.getPidorMessages(ctx, user)
	if err != nil {
		return err
	}

	return s.sendMessages(update.Message.Chat.ID, messages, time.Second, false)
}

func (s *Service) getPidorMessages(ctx context.Context, user domain.User) ([]Message, error) {
	phrase0, phrase1, phrase2, err := s.generatePhrases(ctx)
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

func (s *Service) generatePhrases(ctx context.Context) (phrase0, phrase1, phrase2 string, err error) {
	if !s.enableOpenAI {
		return fallbackPhrases()
	}

	phrase0, phrase1, phrase2, err = s.gptPhrases(ctx)
	if err != nil {
		return fallbackPhrases()
	}
	return phrase0, phrase1, phrase2, nil
}

func fallbackPhrases() (phrase0, phrase1, phrase2 string, err error) {
	phrases := pidorPhrases()
	phrase0, err = randString(phrases[0])
	if err != nil {
		return "", "", "", err
	}
	phrase1, err = randString(phrases[1])
	if err != nil {
		return "", "", "", err
	}
	phrase2, err = randString(phrases[2])
	if err != nil {
		return "", "", "", err
	}
	return phrase0, phrase1, phrase2, nil
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

func containsVotesFromUser(votes []domain.Vote, userID int64) bool {
	for _, vote := range votes {
		if vote.VotedUserID == userID {
			return true
		}
	}
	return false
}
