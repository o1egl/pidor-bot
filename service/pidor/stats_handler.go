package pidor

import (
	"bytes"
	"context"
	"html/template"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/repo"
)

type StatsPeriod string

const (
	StatsPeriodMonth StatsPeriod = "month"
	StatsPeriodYear  StatsPeriod = "year"
	StatsPeriodAll   StatsPeriod = "all"
)

func (s StatsPeriod) TplValue() string {
	switch s {
	case StatsPeriodMonth:
		return "текущий месяц"
	case StatsPeriodYear:
		return "текущий год"
	case StatsPeriodAll:
		return "все время"
	default:
		return "все время"
	}
}

type UserVotes struct {
	Name           string
	Votes          int
	LatestVoteTime time.Time
}

const statsTpl = `
Топ <b>пидоров</b> за {{ .Period }}:

{{ range $index, $user := .Users -}}
<b>{{ add $index 1 }}</b>. {{ $user.Name }} — {{ $user.Votes }} раз(а)
{{ end }}
`

func (s *Service) handleStats(ctx context.Context, update tgbotapi.Update, period StatsPeriod) error {
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

	var filterOpts []repo.StatsFilterOpts
	switch period {
	case StatsPeriodMonth:
		filterOpts = append(filterOpts, repo.WithMonth(int(time.Now().Month())))
	case StatsPeriodYear:
		filterOpts = append(filterOpts, repo.WithYear(time.Now().Year()))
	case StatsPeriodAll:
		// no filter
	}

	votes, err := s.repoClient.GetVotes(ctx, update.Message.Chat.ID, filterOpts...)
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
		"Users":  userVotes,
		"Period": period.TplValue(),
	})
	if err != nil {
		return err
	}

	return s.sendMessage(update.Message.Chat.ID, buf.String())
}

func add(x, y int) int {
	return x + y
}
