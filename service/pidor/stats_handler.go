package pidor

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"text/template"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserVotes struct {
	Name  string
	Votes int
}

const statsTpl = `
Топ *пидоров* за текущий год:

{{ range $index, $user := .Users -}}
*{{ add $index 1 }}*. {{ $user.Name }} — {{ $user.Votes }} раз(а)
{{ end }}
`

func (s *Service) handleStats(ctx context.Context, update tgbotapi.Update) error {
	users, err := s.repoClient.GetUsers(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	votes, err := s.repoClient.GetVotes(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	userVotes := make(map[int64]int)
	for _, vote := range votes {
		userVotes[vote.UserID]++
	}

	userWithVotes := make([]UserVotes, 0, len(users))
	for _, user := range users {
		userWithVotes = append(userWithVotes, UserVotes{
			Name:  strings.TrimPrefix(user.Mention(), "@"),
			Votes: userVotes[user.ID],
		})
	}
	sort.Slice(userWithVotes, func(i, j int) bool {
		return userWithVotes[i].Votes > userWithVotes[j].Votes
	})

	tpl := template.Must(template.New("").Funcs(template.FuncMap{"add": add}).Parse(statsTpl))
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, map[string]interface{}{
		"Users": userWithVotes,
	})
	if err != nil {
		return err
	}

	return s.sendMessage(update.Message.Chat.ID, buf.String())
}

func add(x, y int) int {
	return x + y
}
