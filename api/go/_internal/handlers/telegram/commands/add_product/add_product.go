package add_product

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github/huangc28/kikichoice-be/api/go/_internal/db"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Message constants for better maintainability
const (
	msgStartFlow        = "🆕 開始新的商品上架流程"
	msgNoActiveSession  = "❌ 未找到活動會話"
	msgUnknownOperation = "❌ 未知的操作"
	msgUseAddProduct    = "請使用 /add_product 開始上架商品。"
	msgResumeFlow       = "📋 發現未完成的商品上架流程\n當前步驟: %s\n\n您可以:\n• 繼續輸入以完成當前步驟\n• 輸入 /cancel 取消流程\n• 輸入 /restart 重新開始"
)

// Error message constants
const (
	errMaxImages = "❌ 最多只能上傳 5 張圖片，目前已上傳 %d 張"
)

type AddProductCommand struct {
	dao              *commands.CommandDAO
	productDAO       *ProductDAO
	botAPI           *tgbotapi.BotAPI
	logger           *zap.SugaredLogger
	addProductStates map[string]AddProductState
}

type AddProductCommandParams struct {
	fx.In

	DAO              *commands.CommandDAO
	ProductDAO       *ProductDAO
	BotAPI           *tgbotapi.BotAPI
	Logger           *zap.SugaredLogger
	AddProductStates map[string]AddProductState
}

func NewAddProductCommand(p AddProductCommandParams) *AddProductCommand {
	return &AddProductCommand{
		dao:              p.DAO,
		productDAO:       p.ProductDAO,
		botAPI:           p.BotAPI,
		logger:           p.Logger,
		addProductStates: p.AddProductStates,
	}
}

// Handle processes incoming messages using FSM - simplified for readability
func (c *AddProductCommand) Handle(msg *tgbotapi.Message) error {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID

	state, err := c.getOrCreateUserState(ctx, userID, chatID)
	if err != nil {
		return fmt.Errorf("failed to get user state: %w", err)
	}

	log.Printf("* 1 %+v", state)

	return c.processUserInput(ctx, userID, chatID, state, msg)
}

// processUserInput handles FSM logic - extracted for better readability
func (c *AddProductCommand) processUserInput(ctx context.Context, userID, chatID int64, sessState *AddProductSessionState, msg *tgbotapi.Message) error {
	userFSM := NewAddProductFSM(
		c,
		userID,
		chatID,
		sessState,
		msg,
		c.addProductStates,
	)

	// For new sessions, start the flow
	if sessState.FSMState == StateInit {
		return userFSM.Event(ctx, EventStart)
	}

	availEvents := userFSM.AvailableTransitions()

	if len(availEvents) == 0 {
		return fmt.Errorf("Check your FSM configuration, no available events on current state: %s", sessState.FSMState)
	}

	if err := userFSM.Event(ctx, availEvents[0]); err != nil {
		return fmt.Errorf("FSM event error: %w, current state: %s, event applied: %s", err, sessState.FSMState, availEvents[0])
	}

	return nil
}

// getOrCreateUserState retrieves existing session or creates new one
func (c *AddProductCommand) getOrCreateUserState(ctx context.Context, userID, chatID int64) (*AddProductSessionState, error) {
	session, err := c.dao.GetUserSession(ctx, userID, chatID, "add_product")

	if err == nil {
		var state AddProductSessionState
		if err := json.Unmarshal(session.State, &state); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session state: %w", err)
		}
		return &state, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		state := &AddProductSessionState{
			FSMState:               StateInit,
			Product:                ProductData{},
			Specs:                  []string{},
			ImageFileIDs:           []string{},
			ExpectedReplyMessageID: nil,
		}

		stateJSON, err := json.Marshal(state)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state: %w", err)
		}

		session = &db.UserSession{
			ChatID:      chatID,
			UserID:      userID,
			SessionType: "add_product",
			State:       stateJSON,
		}

		if err := c.dao.UpsertUserSession(ctx, chatID, userID, "add_product", session); err != nil {
			return nil, fmt.Errorf("failed to create user session: %w", err)
		}

		return state, nil
	}

	return nil, err
}

func (c *AddProductCommand) Command() commands.BotCommand {
	return commands.AddProduct
}

var _ commands.CommandHandler = (*AddProductCommand)(nil)
