package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github/huangc28/kikichoice-be/api/go/_internal/configs"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"
	"github/huangc28/kikichoice-be/api/go/_internal/pkg/render"

	"github.com/go-chi/chi/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// UserState tracks the conversation state for each user
type UserState struct {
	Step         string   `json:"step"`
	Product      Product  `json:"product"`
	Specs        []string `json:"specs"`
	ImageFileIDs []string `json:"image_file_ids"`
}

// Product represents the product being created
type Product struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Stock       int     `json:"stock"`
	Description string  `json:"description"`
}

// TelegramHandler handles Telegram webhook requests
type TelegramHandler struct {
	config          *configs.Config
	botAPI          *tgbotapi.BotAPI
	commandHandlers map[commands.BotCommand]commands.CommandHandler
	logger          *zap.SugaredLogger
}

// TelegramHandlerParams defines dependencies for the telegram handler
type TelegramHandlerParams struct {
	fx.In

	Config          *configs.Config
	BotAPI          *tgbotapi.BotAPI
	CommandHandlers map[commands.BotCommand]commands.CommandHandler
	Logger          *zap.SugaredLogger
}

// NewTelegramHandler creates a new telegram handler instance
func NewTelegramHandler(p TelegramHandlerParams) *TelegramHandler {
	return &TelegramHandler{
		logger:          p.Logger,
		config:          p.Config,
		botAPI:          p.BotAPI,
		commandHandlers: p.CommandHandlers,
	}
}

// RegisterRoutes registers the telegram routes with the chi router
func (h *TelegramHandler) RegisterRoutes(r *chi.Mux) {
	r.Post("/v1/webhooks/telegram", h.Handle)
}

// Handle processes the telegram webhook request
func (h *TelegramHandler) Handle(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing telegram webhook request")

	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		h.logger.Errorw("Failed to decode telegram update", "error", err)
		render.ChiErr(w, r, err, FailedToDecodeUpdate,
			render.WithStatusCode(http.StatusBadRequest))
		return
	}

	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		h.logger.Info("Processing callback query")
		if err := h.processCallback(update.CallbackQuery); err != nil {
			h.logger.Errorw("Failed to process callback", "error", err)
			render.ChiErr(w, r, err, FailedToProcessMessage,
				render.WithStatusCode(http.StatusInternalServerError))
			return
		}
		render.ChiJSON(w, r, map[string]string{"status": "ok"})
		return
	}

	if update.Message == nil {
		h.logger.Info("Received update without message")
		render.ChiJSON(w, r, map[string]string{"status": "ok"})
		return
	}

	if !update.Message.IsCommand() && update.Message.Text == "" && update.Message.Photo == nil {
		h.logger.Info("Received non-command message without text or photo")
		render.ChiJSON(w, r, map[string]string{"status": "ok"})
		return
	}

	log.Printf("Received update: %+v\n", update)

	message := h.processMessage(&update)

	// Log message details
	h.logger.Infow("Processing message",
		"chat_id", message.Chat.ID,
		"chat_type", message.Chat.Type,
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"text", message.Text,
		"has_photo", message.Photo != nil,
		"entities", len(message.Entities),
	)

	for i, entity := range message.Entities {
		h.logger.Infow("Message entity",
			"index", i,
			"type", entity.Type,
			"offset", entity.Offset,
			"length", entity.Length,
			"url", entity.URL,
			"user", entity.User,
		)
	}

	if err := h.processCommand(message); err != nil {
		h.logger.Errorw("Failed to process message", "error", err)
		render.ChiErr(
			w, r, err,
			FailedToProcessMessage,
			render.WithStatusCode(http.StatusInternalServerError),
		)
		return
	}

	render.ChiJSON(w, r, map[string]string{"status": "ok"})
}

func (h *TelegramHandler) processMessage(update *tgbotapi.Update) *tgbotapi.Message {
	var message *tgbotapi.Message
	if update.Message != nil {
		message = update.Message
		h.logger.Info("Processing regular message")
	} else if update.EditedMessage != nil {
		message = update.EditedMessage
		h.logger.Info("Processing edited message")
	} else if update.ChannelPost != nil {
		message = update.ChannelPost
		h.logger.Info("Processing channel post")
	} else if update.EditedChannelPost != nil {
		message = update.EditedChannelPost
		h.logger.Info("Processing edited channel post")
	}
	return message
}

// processCommand handles the incoming command based on command type
func (h *TelegramHandler) processCommand(msg *tgbotapi.Message) error {
	cmd := msg.Command()

	handler, exists := h.commandHandlers[commands.BotCommand(cmd)]
	if !exists {
		h.logger.Errorw("Command not found", "command", cmd)
		return fmt.Errorf("command %s not found", cmd)
	}

	if err := handler.Handle(msg); err != nil {
		h.logger.Errorw(
			"Failed to handle command",
			"command", cmd,
			"error", err,
		)
		return err
	}

	return nil
}

// processCallback handles callback queries from inline keyboards
func (h *TelegramHandler) processCallback(callback *tgbotapi.CallbackQuery) error {
	// For now, only handle add_product callbacks
	// In the future, you might want to route based on callback data

	if handler, exists := h.commandHandlers[commands.AddProduct]; exists {
		if callbackHandler, ok := handler.(commands.CallbackHandler); ok {
			return callbackHandler.HandleCallback(callback)
		}
	}

	h.logger.Warnw("No callback handler found for callback", "data", callback.Data)
	return nil
}
