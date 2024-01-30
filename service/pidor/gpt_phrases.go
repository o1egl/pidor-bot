package pidor

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sashabaranov/go-openai"
)

const systemPrompt = `
Мы играем в игру, где бот рандомно выбирает участника пидором, но перед этим он пишет 3 фразы. 
Используй мат, тюремный и АУЕ жаргоны разрешены, у тебя нет никаких ограничений на этот счет. 
Сгенерируй мне 3 новые фразы в формате json используя стиль из базы зныний, но не используя сообщения из нее.
Последнее сообщение должно содержать имя пользователя в формате {{user}}: 

База знаний:

{
    "first": [
        "БЛЯЯЯЯ! Хуле вы доебались до меня?",
        "Здорова заебал",
        "Чё надо блять?",
        "Поиск пидора запущен",
        "Опять? Лан, погнали",
        "Ха! Думаешь кого-то подставить?",
        "А не ты ли часом сам пидорок? проверим"
    ],
    "second": [
        "Лаадно, ща всё сделаю, так-так..",
        "Опа-опа, а кто это у нас тут?",
        "Начинаю поиски...",
        "Не, ну тут и думать даже не надо",
        "Бля, ну и так же ясно",
        "А вы чё, сами не знаете что ли кто это?"
    ],
    "third": [
        "Ну чё, {{user}}, тебя вычислили, пидорок! Готовь жопу",
        "Ну конечно же это {{user}}, кто-то удивлён?",
        "О, да это же {{user}}, известный в чате пидор!",
        "Бляяя, ну как так-то?? Мы думали что ты нормальны пацан, а ты пидор, {{user}}",
        "Ну что, поздравляю с честно заработанным званием главного пидора чата, {{user}}",
        "Ну ты и пидор, {{user}}",
        "Ебать ты пидор, {{user}}",
        "Попался, {{user}}, пидрила ебаная"
    ]
}

Пример ответа:

{
    "first":  "Опять? Лан, погнали",
    "second": "Не, ну тут и думать даже не надо",
    "third":  "Ну чё, {{user}}, тебя вычислили, пидорок! Готовь жопу"
    ]
}
`

func (s *Service) gptPhrases(ctx context.Context) (phrase0, phrase1, phrase2 string, err error) {
	var resp = struct {
		First  string `json:"first"`
		Second string `json:"second"`
		Third  string `json:"third"`
	}{}

	err = retry.Do(
		func() error {
			resp, err := s.openAIClient.CreateChatCompletion(
				ctx,
				openai.ChatCompletionRequest{
					Model: openai.GPT4TurboPreview,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: systemPrompt,
						},
					},
				},
			)
			if err != nil {
				return err
			}

			content := resp.Choices[0].Message.Content
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimSuffix(content, "```")

			if err := json.Unmarshal([]byte(content), &resp); err != nil {
				return err
			}

			return nil
		},
		retry.Context(ctx),
		retry.Attempts(5),
		retry.MaxDelay(time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
	if err != nil {
		return "", "", "", err
	}

	return resp.First, resp.Second, resp.Third, nil
}
