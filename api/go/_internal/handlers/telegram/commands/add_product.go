package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
	"go.uber.org/fx"
)

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

// FSM Events
const (
	EventStart   = "start"
	EventNext    = "next"
	EventSkip    = "skip"
	EventDone    = "done"
	EventCancel  = "cancel"
	EventRestart = "restart"
	EventConfirm = "confirm"
	EventReject  = "reject"
	EventPause   = "pause"
	EventResume  = "resume"
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

// FSMContext holds context for FSM callbacks
type FSMContext struct {
	UserID  int64
	ChatID  int64
	Message *tgbotapi.Message
	State   *UserState
	Command *AddProductCommand
}

func NewAddProductCommand(p AddProductCommandParams) *AddProductCommand {
	return &AddProductCommand{
		dao:        p.DAO,
		productDAO: p.ProductDAO,
		botAPI:     p.BotAPI,
	}
}

// Handle processes incoming messages using FSM
func (c *AddProductCommand) Handle(msg *tgbotapi.Message) error {
	ctx := context.Background()
	userID := msg.From.ID
	chatID := msg.Chat.ID
	text := msg.Text

	// Get or create user state
	state, err := c.getOrCreateUserState(ctx, userID, chatID, text)
	if err != nil {
		return fmt.Errorf("failed to get user state: %w", err)
	}

	if state == nil {
		return c.sendMessage(chatID, "請使用 /add_product 開始上架商品。")
	}

	// Create FSM instance
	userFSM := c.createFSM(userID, chatID, state, msg)

	// Set current FSM state
	userFSM.SetState(state.FSMState)

	// Determine event based on input and current state
	event := c.determineEvent(text, userFSM.Current(), msg)

	// Store input for validation
	state.CurrentInput = text

	// Trigger FSM event
	if err := userFSM.Event(ctx, event); err != nil {
		// Handle FSM errors (invalid transitions, validation errors, etc.)
		if err.Error() == "event "+event+" inappropriate in current state "+userFSM.Current() {
			return c.handleInvalidInput(chatID, userFSM.Current(), text)
		}
		// Handle validation errors
		if err.Error() == "invalid price format" || err.Error() == "invalid stock format" {
			return c.handleInvalidInput(chatID, userFSM.Current(), text)
		}
		if err.Error() == "maximum images reached" {
			return c.sendMessage(chatID, fmt.Sprintf("❌ 最多只能上傳 5 張圖片，目前已上傳 %d 張", len(state.ImageFileIDs)))
		}
		return fmt.Errorf("FSM event error: %w", err)
	}

	// Update FSM state in user state
	state.FSMState = userFSM.Current()

	// Save updated state (only if not completed or cancelled)
	if state.FSMState != StateCompleted && state.FSMState != StateCancelled {
		if err := c.dao.UpdateUserSession(ctx, userID, "add_product", state); err != nil {
			return fmt.Errorf("failed to save user state: %w", err)
		}
	}

	return nil
}

// createFSM creates a new FSM instance with all events and callbacks
func (c *AddProductCommand) createFSM(userID, chatID int64, state *UserState, msg *tgbotapi.Message) *fsm.FSM {
	fsmCtx := &FSMContext{
		UserID:  userID,
		ChatID:  chatID,
		Message: msg,
		State:   state,
		Command: c,
	}

	return fsm.NewFSM(
		StateInit,
		fsm.Events{
			// Start flow
			{Name: EventStart, Src: []string{StateInit}, Dst: StateSKU},

			// Normal progression
			{Name: EventNext, Src: []string{StateSKU}, Dst: StateName},
			{Name: EventNext, Src: []string{StateName}, Dst: StateCategory},
			{Name: EventNext, Src: []string{StateCategory}, Dst: StatePrice},
			{Name: EventNext, Src: []string{StatePrice}, Dst: StateStock},
			{Name: EventNext, Src: []string{StateStock}, Dst: StateDescription},
			{Name: EventNext, Src: []string{StateDescription}, Dst: StateSpecs},
			{Name: EventNext, Src: []string{StateSpecs}, Dst: StateSpecs},   // Stay in specs for multiple entries
			{Name: EventNext, Src: []string{StateImages}, Dst: StateImages}, // Stay in images for multiple uploads

			// Skip optional states
			{Name: EventSkip, Src: []string{StateDescription}, Dst: StateSpecs},
			{Name: EventSkip, Src: []string{StateSpecs}, Dst: StateImages},
			{Name: EventSkip, Src: []string{StateImages}, Dst: StateConfirm},

			// Done events for multi-input states
			{Name: EventDone, Src: []string{StateSpecs}, Dst: StateImages},
			{Name: EventDone, Src: []string{StateImages}, Dst: StateConfirm},

			// Confirmation
			{Name: EventConfirm, Src: []string{StateConfirm}, Dst: StateCompleted},
			{Name: EventReject, Src: []string{StateConfirm}, Dst: StateCancelled},

			// Global events
			{Name: EventCancel, Src: []string{"*"}, Dst: StateCancelled},
			{Name: EventRestart, Src: []string{"*"}, Dst: StateSKU},
			{Name: EventPause, Src: []string{"*"}, Dst: StatePaused},
			{Name: EventResume, Src: []string{"*"}, Dst: "*"}, // Resume from where left off
		},
		fsm.Callbacks{
			// Enter state callbacks (send prompts)
			"enter_" + StateSKU:         func(ctx context.Context, e *fsm.Event) { c.enterSKU(ctx, e, fsmCtx) },
			"enter_" + StateName:        func(ctx context.Context, e *fsm.Event) { c.enterName(ctx, e, fsmCtx) },
			"enter_" + StateCategory:    func(ctx context.Context, e *fsm.Event) { c.enterCategory(ctx, e, fsmCtx) },
			"enter_" + StatePrice:       func(ctx context.Context, e *fsm.Event) { c.enterPrice(ctx, e, fsmCtx) },
			"enter_" + StateStock:       func(ctx context.Context, e *fsm.Event) { c.enterStock(ctx, e, fsmCtx) },
			"enter_" + StateDescription: func(ctx context.Context, e *fsm.Event) { c.enterDescription(ctx, e, fsmCtx) },
			"enter_" + StateSpecs:       func(ctx context.Context, e *fsm.Event) { c.enterSpecs(ctx, e, fsmCtx) },
			"enter_" + StateImages:      func(ctx context.Context, e *fsm.Event) { c.enterImages(ctx, e, fsmCtx) },
			"enter_" + StateConfirm:     func(ctx context.Context, e *fsm.Event) { c.enterConfirm(ctx, e, fsmCtx) },
			"enter_" + StateCompleted:   func(ctx context.Context, e *fsm.Event) { c.enterCompleted(ctx, e, fsmCtx) },
			"enter_" + StateCancelled:   func(ctx context.Context, e *fsm.Event) { c.enterCancelled(ctx, e, fsmCtx) },
			"enter_" + StatePaused:      func(ctx context.Context, e *fsm.Event) { c.enterPaused(ctx, e, fsmCtx) },

			// Before event callbacks (validation)
			"before_" + EventNext: func(ctx context.Context, e *fsm.Event) { c.validateInput(ctx, e, fsmCtx) },

			// After event callbacks (data storage)
			"after_" + EventNext: func(ctx context.Context, e *fsm.Event) { c.storeInput(ctx, e, fsmCtx) },
		},
	)
}

// determineEvent maps user input to FSM events
func (c *AddProductCommand) determineEvent(text, currentState string, msg *tgbotapi.Message) string {
	// Handle global commands
	switch text {
	case "/cancel":
		return EventCancel
	case "/restart":
		return EventRestart
	case "/add_product":
		if currentState == StateInit {
			return EventStart
		}
		return EventResume
	case "/done":
		if currentState == StateSpecs || currentState == StateImages {
			return EventDone
		}
	}

	// Handle confirmation
	if currentState == StateConfirm {
		if text == "確認" {
			return EventConfirm
		} else if text == "取消" {
			return EventReject
		}
	}

	// Handle image uploads
	if currentState == StateImages && msg.Photo != nil {
		return EventNext
	}

	// Default to next for text input
	return EventNext
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
			c.sendMessage(chatID, "🆕 開始新的商品上架流程")
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
		resumeMsg := fmt.Sprintf("📋 發現未完成的商品上架流程\n當前步驟: %s\n\n您可以:\n• 繼續輸入以完成當前步驟\n• 輸入 /cancel 取消流程\n• 輸入 /restart 重新開始", currentStepMsg)
		c.sendMessage(chatID, resumeMsg)
	}

	return &state, nil
}

// FSM State Entry Callbacks

func (c *AddProductCommand) enterSKU(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "請輸入商品 SKU：")
}

func (c *AddProductCommand) enterName(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "請輸入商品名稱：")
}

func (c *AddProductCommand) enterCategory(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "請輸入商品類別：")
}

func (c *AddProductCommand) enterPrice(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "請輸入商品價格：")
}

func (c *AddProductCommand) enterStock(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "請輸入商品庫存數量：")
}

func (c *AddProductCommand) enterDescription(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, "請輸入商品描述：", "description")
}

func (c *AddProductCommand) enterSpecs(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, "請輸入商品規格（每行一項）：", "specs")
}

func (c *AddProductCommand) enterImages(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, "請上傳商品圖片（最多 5 張）：", "images")
}

func (c *AddProductCommand) enterConfirm(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendSummary(fsmCtx.ChatID, fsmCtx.State)
}

func (c *AddProductCommand) enterCompleted(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	if err := c.productDAO.SaveProduct(ctx, fsmCtx.State); err != nil {
		c.sendMessage(fsmCtx.ChatID, "❌ 儲存失敗："+err.Error())
	} else {
		c.sendMessage(fsmCtx.ChatID, "🎉 商品已成功上架！")
	}
	// Clean up session
	c.dao.DeleteUserSession(ctx, fsmCtx.UserID, "add_product")
}

func (c *AddProductCommand) enterCancelled(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "❌ 已取消商品上架流程")
	// Clean up session
	c.dao.DeleteUserSession(ctx, fsmCtx.UserID, "add_product")
}

func (c *AddProductCommand) enterPaused(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, "💾 流程已暫存，您可以稍後使用 /add_product 繼續")
}

// FSM Event Callbacks

// validateInput checks if input is valid before proceeding with FSM event
func (c *AddProductCommand) validateInput(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	switch e.Src {
	case StatePrice:
		if _, err := strconv.ParseFloat(fsmCtx.State.CurrentInput, 64); err != nil {
			// Return error to prevent state transition
			e.Cancel(fmt.Errorf("invalid price format"))
			return
		}
	case StateStock:
		if _, err := strconv.Atoi(fsmCtx.State.CurrentInput); err != nil {
			e.Cancel(fmt.Errorf("invalid stock format"))
			return
		}
	case StateImages:
		if fsmCtx.Message != nil && fsmCtx.Message.Photo != nil && len(fsmCtx.State.ImageFileIDs) >= 5 {
			e.Cancel(fmt.Errorf("maximum images reached"))
			return
		}
	}
}

func (c *AddProductCommand) storeInput(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	switch e.Src {
	case StateSKU:
		fsmCtx.State.Product.SKU = fsmCtx.State.CurrentInput
	case StateName:
		fsmCtx.State.Product.Name = fsmCtx.State.CurrentInput
	case StateCategory:
		fsmCtx.State.Product.Category = fsmCtx.State.CurrentInput
	case StatePrice:
		price, _ := strconv.ParseFloat(fsmCtx.State.CurrentInput, 64)
		fsmCtx.State.Product.Price = price
	case StateStock:
		stock, _ := strconv.Atoi(fsmCtx.State.CurrentInput)
		fsmCtx.State.Product.Stock = stock
	case StateDescription:
		fsmCtx.State.Product.Description = fsmCtx.State.CurrentInput
	case StateSpecs:
		if fsmCtx.State.CurrentInput != "/done" {
			fsmCtx.State.Specs = append(fsmCtx.State.Specs, fsmCtx.State.CurrentInput)
			// Send feedback for specs
			c.sendMessage(fsmCtx.ChatID, "✅ 規格已新增，繼續輸入或點擊「完成」按鈕：")
		}
	case StateImages:
		if fsmCtx.Message != nil && fsmCtx.Message.Photo != nil {
			fileID := fsmCtx.Message.Photo[len(fsmCtx.Message.Photo)-1].FileID
			fsmCtx.State.ImageFileIDs = append(fsmCtx.State.ImageFileIDs, fileID)

			// Send feedback for images
			const maxImages = 5
			remaining := maxImages - len(fsmCtx.State.ImageFileIDs)
			if remaining > 0 {
				c.sendMessage(fsmCtx.ChatID, fmt.Sprintf("✅ 圖片已上傳 (%d/%d)，還可上傳 %d 張或點擊「完成」按鈕", len(fsmCtx.State.ImageFileIDs), maxImages, remaining))
			} else {
				c.sendMessage(fsmCtx.ChatID, fmt.Sprintf("✅ 圖片已上傳 (%d/%d)，已達上限！點擊「完成」按鈕", len(fsmCtx.State.ImageFileIDs), maxImages))
			}
		}
	}
}

// handleInvalidInput handles invalid input for current state
func (c *AddProductCommand) handleInvalidInput(chatID int64, currentState, input string) error {
	switch currentState {
	case StatePrice:
		return c.sendMessage(chatID, "❌ 價格格式錯誤，請輸入數字：")
	case StateStock:
		return c.sendMessage(chatID, "❌ 庫存格式錯誤，請輸入整數：")
	case StateImages:
		return c.sendMessage(chatID, fmt.Sprintf("❌ 最多只能上傳 5 張圖片，目前已上傳 %d 張", 5))
	default:
		return c.sendMessage(chatID, "❌ 輸入格式錯誤，請重新輸入：")
	}
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
		return c.sendMessage(chatID, "❌ 未找到活動會話")
	}

	var state UserState
	if err := json.Unmarshal(session.State, &state); err != nil {
		return err
	}

	// Create FSM instance and set current state
	userFSM := c.createFSM(userID, chatID, &state, nil)
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
		return c.sendMessage(chatID, "❌ 未知的操作")
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
