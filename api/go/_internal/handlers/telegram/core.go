package telegram

import (
	"github/huangc28/kikichoice-be/api/go/_internal/configs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBotAPI(cfg *configs.Config) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, err
	}
	return bot, nil
}
