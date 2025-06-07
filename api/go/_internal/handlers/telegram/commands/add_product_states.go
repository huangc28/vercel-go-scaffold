package commands

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// FSM States
const (
	StateInit        = "init"
	StateSKU         = "sku"
	StateName        = "name"
	StateCategory    = "category"
	StatePrice       = "price"
	StateStock       = "stock"
	StateDescription = "description"
	StateSpecs       = "specs"
	StateImages      = "images"
	StateConfirm     = "confirm"
	StateCompleted   = "completed"
	StateCancelled   = "cancelled"
	StatePaused      = "paused"
)

type AddProductState interface {
	Name() string
	Buttons() []tgbotapi.InlineKeyboardButton
}

type AddProductStateInit struct{}

func (s *AddProductStateInit) Name() string {
	return "init"
}

func (s *AddProductStateInit) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{}
}
