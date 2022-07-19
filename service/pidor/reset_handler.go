package pidor

import (
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/repo"
)

const (
	resetDay   = "day"
	resetMonth = "month"
	resetYear  = "year"
	resetAll   = "all"
)

func (s *Service) handleReset(ctx context.Context, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	if update.Message.From.UserName != s.adminUsername {
		return s.sendMessage(chatID, "Пошел нахуй пидарасина!")
	}

	var resetPeriod string
	if parts := strings.Fields(update.Message.Text); len(parts) > 1 {
		resetPeriod = parts[1]
	}

	resetFilters := make([]repo.StatsFilterOpts, 0)
	switch resetPeriod {
	case resetDay:
		resetFilters = append(resetFilters, repo.WithDay(time.Now().Day()))
	case resetMonth:
		resetFilters = append(resetFilters, repo.WithMonth(int(time.Now().Month())))
	case resetYear:
		resetFilters = append(resetFilters, repo.WithYear(time.Now().Year()))
	case resetAll:
	default:
		return s.sendMessage(chatID, "Неизвестный период для обнуления")
	}

	return s.repoClient.DeleteVotes(ctx, chatID, resetFilters...)
}
