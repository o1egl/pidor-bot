package pidor

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/o1egl/pidor-bot/runes"
)

type MessageVarType string

const (
	MessageVarTypeText        MessageVarType = "text"
	MessageVarTypeTextMention MessageVarType = "text_mention"
)

type MessageVar struct {
	Name  string
	Value string
	Type  MessageVarType
	User  *tgbotapi.User
}

func NewMentionVar(varName, value string, userID int64) MessageVar {
	return MessageVar{
		Name:  varName,
		Value: value,
		Type:  MessageVarTypeTextMention,
		User:  &tgbotapi.User{ID: userID},
	}
}

func (s *Service) sendMessage(chatID int64, text string, vars ...MessageVar) error {
	var messageEntities []tgbotapi.MessageEntity
	for _, variable := range vars {
		var (
			offset int
			length int
			err    error
		)

		text, offset, length, err = renderTpl(text, variable.Name, variable.Value)
		if err != nil {
			return err
		}
		switch variable.Type {
		case MessageVarTypeText:
		case MessageVarTypeTextMention:
			messageEntities = append(messageEntities, tgbotapi.MessageEntity{
				Type:   "text_mention",
				Offset: offset,
				Length: length,
				User:   variable.User,
			})
		}
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.Entities = messageEntities
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := s.bot.Send(msg)
	return err
}

func renderTpl(tpl string, variable string, value string) (_ string, offset, length int, _ error) {
	tplRunes := []rune(tpl)
	varRunes := []rune(variable)
	offset = runes.Index(tplRunes, varRunes)
	if offset == -1 {
		return "", 0, 0, fmt.Errorf("variable %s not found", variable)
	}

	length = len([]rune(value))

	return strings.Replace(tpl, variable, value, 1), offset, length, nil
}
