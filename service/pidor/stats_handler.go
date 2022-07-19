package pidor

import (
	"bytes"
	"context"
	"html/template"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserVotes struct {
	Name           string
	Votes          int
	LatestVoteTime time.Time
}

const statsTpl = `
Топ <b>пидоров</b> за текущий год:

{{ range $index, $user := .Users -}}
<b>{{ add $index 1 }}</b>. {{ $user.Name }} — {{ $user.Votes }} раз(а)
{{ end }}
`

func (s *Service) handleStats(ctx context.Context, update tgbotapi.Update) error {
	users, err := s.repoClient.GetUsers(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}

	userVotesMap := make(map[int64]UserVotes)
	for _, user := range users {
		userVotesMap[user.ID] = UserVotes{
			Name: strings.TrimPrefix(user.Mention(), "@"),
		}
	}

	votes, err := s.repoClient.GetVotes(ctx, update.Message.Chat.ID)
	if err != nil {
		return err
	}
	for _, vote := range votes {
		userVotes, ok := userVotesMap[vote.UserID]
		if ok {
			userVotes.Votes++
			userVotes.LatestVoteTime = vote.Time
			userVotesMap[vote.UserID] = userVotes
		}
	}

	userVotes := make([]UserVotes, 0, len(userVotesMap))
	for _, userVote := range userVotesMap {
		userVotes = append(userVotes, userVote)
	}

	sort.Slice(userVotes, func(i, j int) bool {
		userI := userVotes[i]
		userJ := userVotes[j]
		if userI.Votes == userJ.Votes {
			if userI.LatestVoteTime.Equal(userJ.LatestVoteTime) {
				return userI.Name < userJ.Name
			}
			return userI.LatestVoteTime.After(userJ.LatestVoteTime)
		}
		return userI.Votes > userJ.Votes
	})

	tpl := template.Must(template.New("").Funcs(template.FuncMap{"add": add}).Parse(statsTpl))
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, map[string]interface{}{
		"Users": userVotes,
	})
	if err != nil {
		return err
	}

	return s.sendMessage(update.Message.Chat.ID, buf.String())
}

func add(x, y int) int {
	return x + y
}
