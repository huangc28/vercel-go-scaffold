package telegram

import (
	"context"
	"encoding/json"
	"fmt"
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
	replyProcessor  *ReplyProcessor
	logger          *zap.SugaredLogger
}

// TelegramHandlerParams defines dependencies for the telegram handler
type TelegramHandlerParams struct {
	fx.In

	Config          *configs.Config
	BotAPI          *tgbotapi.BotAPI
	CommandHandlers map[commands.BotCommand]commands.CommandHandler
	ReplyProcessor  *ReplyProcessor
	Logger          *zap.SugaredLogger
}

// NewTelegramHandler creates a new telegram handler instance
func NewTelegramHandler(p TelegramHandlerParams) *TelegramHandler {
	return &TelegramHandler{
		logger:          p.Logger,
		config:          p.Config,
		botAPI:          p.BotAPI,
		commandHandlers: p.CommandHandlers,
		replyProcessor:  p.ReplyProcessor,
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

	if update.Message == nil {
		h.logger.Info("Received update without message")
		render.ChiJSON(w, r, nil)
		return
	}

	message := h.retrieveMessage(&update)
	if err := h.processMessage(r.Context(), message); err != nil {
		h.logger.Errorw("Failed to process message", "error", err)
		render.ChiErr(
			w, r, err,
			FailedToProcessMessage,
			render.WithStatusCode(http.StatusOK),
		)
		return
	}

	render.ChiJSON(w, r, nil)
}

func (h *TelegramHandler) retrieveMessage(update *tgbotapi.Update) *tgbotapi.Message {
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

// processMessage handles the incoming message based on message type
func (h *TelegramHandler) processMessage(ctx context.Context, msg *tgbotapi.Message) error {
	if h.isReplyToCommand(msg) {
		if err := h.replyProcessor.Process(ctx, msg); err != nil {
			h.logger.Errorw("Failed to process reply", "error", err)
			return err
		}
		return nil
	}

	if msg.IsCommand() {
		h.logger.Infow("Processing command", "command", msg.Command())

		handler, exists := h.commandHandlers[commands.BotCommand(msg.Command())]
		if !exists {
			h.logger.Errorw("Command not found", "command", msg.Command())
			return fmt.Errorf("command %s not found", msg.Command())
		}

		if err := handler.Handle(msg); err != nil {
			h.logger.Errorw(
				"Failed to handle command",
				"command", msg.Command(),
				"error", err,
			)
			return err
		}
	}

	return nil
}

func (h *TelegramHandler) isReplyToCommand(msg *tgbotapi.Message) bool {
	return msg.ReplyToMessage != nil
}
