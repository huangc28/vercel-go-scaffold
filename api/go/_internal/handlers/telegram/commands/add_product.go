package commands

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
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
	dao        *CommandDAO
	productDAO *ProductDAO
	botAPI     *tgbotapi.BotAPI
}

type AddProductCommandParams struct {
	fx.In

	DAO        *CommandDAO
	ProductDAO *ProductDAO
	BotAPI     *tgbotapi.BotAPI
}

// UserState represents the product data and current input
type UserState struct {
	Product      ProductData `json:"product"`
	Specs        []string    `json:"specs"`
	ImageFileIDs []string    `json:"image_file_ids"`
	CurrentInput string      `json:"current_input"`
	FSMState     string      `json:"fsm_state"`
}

type ProductData struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Description string  `json:"description"`
}

func NewAddProductCommand(p AddProductCommandParams) *AddProductCommand {
	return &AddProductCommand{
		dao:        p.DAO,
		productDAO: p.ProductDAO,
		botAPI:     p.BotAPI,
	}
}

// Handle processes incoming messages using FSM - simplified for readability
func (c *AddProductCommand) Handle(msg *tgbotapi.Message) error {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	state, err := c.getOrCreateUserState(ctx, userID, chatID, text)
	if err != nil {
		return fmt.Errorf("failed to get user state: %w", err)
	}

	if state == nil {
		return c.sendMessage(chatID, msgUseAddProduct)
	}

	return c.processUserInput(ctx, userID, chatID, state, msg)
}

// processUserInput handles FSM logic - extracted for better readability
func (c *AddProductCommand) processUserInput(ctx context.Context, userID, chatID int64, state *UserState, msg *tgbotapi.Message) error {
	userFSM := NewAddProductFSM(c, userID, chatID, state, msg)
	userFSM.SetState(state.FSMState)

	event := c.determineEvent(msg.Text, userFSM.Current(), msg)
	state.CurrentInput = msg.Text

	if err := userFSM.Event(ctx, event); err != nil {
		return c.handleFSMError(err, chatID, userFSM.Current(), msg.Text, state, event)
	}

	state.FSMState = userFSM.Current()
	return c.saveStateIfNeeded(ctx, userID, state)
}

// handleFSMError provides centralized error handling
func (c *AddProductCommand) handleFSMError(err error, chatID int64, currentState, input string, state *UserState, event string) error {
	switch err.Error() {
	case "invalid price format", "invalid stock format":
		return c.handleInvalidInput(chatID, currentState, input)
	case "maximum images reached":
		return c.sendMessage(chatID, fmt.Sprintf(errMaxImages, len(state.ImageFileIDs)))
	default:
		if err.Error() == "event "+event+" inappropriate in current state "+currentState {
			return c.handleInvalidInput(chatID, currentState, input)
		}
		return fmt.Errorf("FSM event error: %w", err)
	}
}

// saveStateIfNeeded handles state persistence logic
func (c *AddProductCommand) saveStateIfNeeded(ctx context.Context, userID int64, state *UserState) error {
	if state.FSMState != StateCompleted && state.FSMState != StateCancelled {
		if err := c.dao.UpdateUserSession(ctx, userID, "add_product", state); err != nil {
			return fmt.Errorf("failed to save user state: %w", err)
		}
	}
	return nil
}

// getOrCreateUserState retrieves existing session or creates new one
func (c *AddProductCommand) getOrCreateUserState(ctx context.Context, userID int64, chatID int64, text string) (*UserState, error) {
	// Try to get existing session
	session, err := c.dao.GetUserSession(ctx, userID, "add_product")
	if err != nil {
		return nil, fmt.Errorf("failed to get user session: %w", err)
	}

	if session == nil {
		// If no session exists and command is /add_product, create new one
		if text == "/add_product" {
			state := &UserState{FSMState: StateInit}
			if err := c.dao.CreateUserSession(ctx, chatID, userID, "add_product", state); err != nil {
				return nil, fmt.Errorf("failed to create user session: %w", err)
			}
			c.sendMessage(chatID, msgStartFlow)
			return state, nil
		}
		return nil, nil
	}

	// Parse existing session state
	var state UserState
	if err := json.Unmarshal(session.State, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session state: %w", err)
	}

	// Handle resume for existing session
	if text == "/add_product" {
		currentStepMsg := c.getStepDescription(state.FSMState)
		resumeMsg := fmt.Sprintf(msgResumeFlow, currentStepMsg)
		c.sendMessage(chatID, resumeMsg)
	}

	return &state, nil
}

// sendMessage sends a text message to the chat
func (c *AddProductCommand) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.botAPI.Send(msg)
	return err
}

// sendMessageWithButtons sends a message with inline keyboard buttons
func (c *AddProductCommand) sendMessageWithButtons(chatID int64, text string, step string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	// Create keyboard based on step type
	var keyboard tgbotapi.InlineKeyboardMarkup

	if c.canSkipStep(step) {
		if c.needsDoneButton(step) {
			// For steps that need a "Done" button (specs, images)
			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✅ 完成", fmt.Sprintf("done_%s", step)),
					tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", fmt.Sprintf("skip_%s", step)),
				),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
					tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
				),
			)
		} else {
			// For other skippable steps (description)
			keyboard = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
					tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", fmt.Sprintf("skip_%s", step)),
					tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
				),
			)
		}
		msg.ReplyMarkup = keyboard
	}

	_, err := c.botAPI.Send(msg)
	return err
}

// canSkipStep determines if a step can be skipped
func (c *AddProductCommand) canSkipStep(step string) bool {
	skippableSteps := map[string]bool{
		"description": true,
		"specs":       true,
		"images":      true,
	}
	return skippableSteps[step]
}

// needsDoneButton determines if a step needs a "Done" button
func (c *AddProductCommand) needsDoneButton(step string) bool {
	stepsWithDone := map[string]bool{
		"specs":  true,
		"images": true,
	}
	return stepsWithDone[step]
}

// sendSummary sends a product summary for confirmation
func (c *AddProductCommand) sendSummary(chatID int64, state *UserState) error {
	summary := fmt.Sprintf(
		"商品摘要：\nSKU: %s\n名稱: %s\n類別: %s\n價格: %.2f\n庫存: %d\n描述: %s\n規格: %v\n圖片數量: %d\n請選擇：",
		state.Product.SKU,
		state.Product.Name,
		state.Product.Category,
		state.Product.Price,
		state.Product.Stock,
		state.Product.Description,
		state.Specs,
		len(state.ImageFileIDs),
	)

	msg := tgbotapi.NewMessage(chatID, summary)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ 確認", "confirm"),
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := c.botAPI.Send(msg)
	return err
}

// getStepDescription returns a user-friendly description of the current step
func (c *AddProductCommand) getStepDescription(state string) string {
	descriptions := map[string]string{
		StateSKU:         "輸入商品 SKU",
		StateName:        "輸入商品名稱",
		StateCategory:    "輸入商品類別",
		StatePrice:       "輸入商品價格",
		StateStock:       "輸入商品庫存數量",
		StateDescription: "輸入商品描述",
		StateSpecs:       "輸入商品規格",
		StateImages:      "上傳商品圖片",
		StateConfirm:     "確認商品資訊",
	}

	if desc, exists := descriptions[state]; exists {
		return desc
	}
	return "未知步驟"
}

// HandleCallback handles inline keyboard button presses
func (c *AddProductCommand) HandleCallback(callback *tgbotapi.CallbackQuery) error {
	ctx := context.Background()
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Get current user state
	session, err := c.dao.GetUserSession(ctx, userID, "add_product")
	if err != nil || session == nil {
		return c.sendMessage(chatID, msgNoActiveSession)
	}

	var state UserState
	if err := json.Unmarshal(session.State, &state); err != nil {
		return err
	}

	// Create FSM instance and set current state
	userFSM := NewAddProductFSM(c, userID, chatID, &state, nil)
	userFSM.SetState(state.FSMState)

	// Map callback data to FSM events
	var event string
	switch {
	case data == "cancel":
		event = EventCancel
	case data == "confirm":
		event = EventConfirm
	case data == "pause":
		event = EventPause
	case len(data) > 5 && data[:5] == "skip_":
		event = EventSkip
	case len(data) > 5 && data[:5] == "done_":
		event = EventDone
	default:
		return c.sendMessage(chatID, msgUnknownOperation)
	}

	// Trigger FSM event
	if err := userFSM.Event(ctx, event); err != nil {
		return fmt.Errorf("FSM callback event error: %w", err)
	}

	// Update FSM state
	state.FSMState = userFSM.Current()

	// Save updated state (only if not completed or cancelled)
	if state.FSMState != StateCompleted && state.FSMState != StateCancelled {
		if err := c.dao.UpdateUserSession(ctx, userID, "add_product", &state); err != nil {
			return fmt.Errorf("failed to save user state: %w", err)
		}
	}

	return nil
}

func (c *AddProductCommand) Command() BotCommand {
	return AddProduct
}

var _ CommandHandler = (*AddProductCommand)(nil)
