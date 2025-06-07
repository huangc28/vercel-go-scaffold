package commands

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
)

// UI Message constants for FSM states
const (
	promptSKU         = "請輸入商品 SKU："
	promptName        = "請輸入商品名稱："
	promptCategory    = "請輸入商品類別："
	promptPrice       = "請輸入商品價格："
	promptStock       = "請輸入商品庫存數量："
	promptDescription = "請輸入商品描述："
	promptSpecs       = "請輸入商品規格（每行一項）："
	promptImages      = "請上傳商品圖片（最多 5 張）："

	msgSuccess           = "🎉 商品已成功上架！"
	msgCancelled         = "❌ 已取消商品上架流程"
	msgPaused            = "💾 流程已暫存，您可以稍後使用 /add_product 繼續"
	msgSpecAdded         = "✅ 規格已新增，繼續輸入或點擊「完成」按鈕："
	msgImageUploaded     = "✅ 圖片已上傳 (%d/%d)，還可上傳 %d 張或點擊「完成」按鈕"
	msgImageLimitReached = "✅ 圖片已上傳 (%d/%d)，已達上限！點擊「完成」按鈕"

	msgInvalidPrice = "❌ 價格格式錯誤，請輸入數字："
	msgInvalidStock = "❌ 庫存格式錯誤，請輸入整數："
	msgInvalidInput = "❌ 輸入格式錯誤，請重新輸入："
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

// FSMContext holds context for FSM callbacks
type FSMContext struct {
	UserID  int64
	ChatID  int64
	Message *tgbotapi.Message
	State   *UserState
	Command *AddProductCommand
}

// NewAddProductFSM creates a new FSM instance with all events and callbacks
func NewAddProductFSM(c *AddProductCommand, userID, chatID int64, state *UserState, msg *tgbotapi.Message) *fsm.FSM {
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

// FSM State Entry Callbacks

func (c *AddProductCommand) enterSKU(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, promptSKU)
}

func (c *AddProductCommand) enterName(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptName, StateName)
}

func (c *AddProductCommand) enterCategory(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptCategory, StateCategory)
}

func (c *AddProductCommand) enterPrice(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptPrice, StatePrice)
}

func (c *AddProductCommand) enterStock(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptStock, StateStock)
}

func (c *AddProductCommand) enterDescription(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptDescription, StateDescription)
}

func (c *AddProductCommand) enterSpecs(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptSpecs, StateSpecs)
}

func (c *AddProductCommand) enterImages(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessageWithButtons(fsmCtx.ChatID, promptImages, StateImages)
}

func (c *AddProductCommand) enterConfirm(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendSummary(fsmCtx.ChatID, fsmCtx.State)
}

func (c *AddProductCommand) enterCompleted(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	if err := c.productDAO.SaveProduct(ctx, fsmCtx.State); err != nil {
		c.sendMessage(fsmCtx.ChatID, "❌ 儲存失敗："+err.Error())
	} else {
		c.sendMessage(fsmCtx.ChatID, msgSuccess)
	}
	// Clean up session
	c.dao.DeleteUserSession(ctx, fsmCtx.UserID, "add_product")
}

func (c *AddProductCommand) enterCancelled(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, msgCancelled)
	// Clean up session
	c.dao.DeleteUserSession(ctx, fsmCtx.UserID, "add_product")
}

func (c *AddProductCommand) enterPaused(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	c.sendMessage(fsmCtx.ChatID, msgPaused)
}

// FSM Event Callbacks

// validateInput checks if input is valid before proceeding with FSM event
func (c *AddProductCommand) validateInput(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	switch e.Src {
	case StatePrice:
		if !c.isValidPrice(fsmCtx.State.CurrentInput) {
			e.Cancel(fmt.Errorf("invalid price format"))
			return
		}
	case StateStock:
		if !c.isValidStock(fsmCtx.State.CurrentInput) {
			e.Cancel(fmt.Errorf("invalid stock format"))
			return
		}
	case StateImages:
		if !c.isValidImageUpload(fsmCtx) {
			e.Cancel(fmt.Errorf("maximum images reached"))
			return
		}
	}
}

// Validation helper methods
func (c *AddProductCommand) isValidPrice(input string) bool {
	_, err := strconv.ParseFloat(input, 64)
	return err == nil
}

func (c *AddProductCommand) isValidStock(input string) bool {
	_, err := strconv.Atoi(input)
	return err == nil
}

func (c *AddProductCommand) isValidImageUpload(fsmCtx *FSMContext) bool {
	return !(fsmCtx.Message != nil && fsmCtx.Message.Photo != nil && len(fsmCtx.State.ImageFileIDs) >= 5)
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
			c.sendMessage(fsmCtx.ChatID, msgSpecAdded)
		}
	case StateImages:
		if fsmCtx.Message != nil && fsmCtx.Message.Photo != nil {
			fileID := fsmCtx.Message.Photo[len(fsmCtx.Message.Photo)-1].FileID
			fsmCtx.State.ImageFileIDs = append(fsmCtx.State.ImageFileIDs, fileID)

			// Send feedback for images
			const maxImages = 5
			remaining := maxImages - len(fsmCtx.State.ImageFileIDs)
			if remaining > 0 {
				c.sendMessage(fsmCtx.ChatID, fmt.Sprintf(msgImageUploaded, len(fsmCtx.State.ImageFileIDs), maxImages, remaining))
			} else {
				c.sendMessage(fsmCtx.ChatID, fmt.Sprintf(msgImageLimitReached, len(fsmCtx.State.ImageFileIDs), maxImages))
			}
		}
	}
}

// handleInvalidInput handles invalid input for current state
func (c *AddProductCommand) handleInvalidInput(chatID int64, currentState, input string) error {
	switch currentState {
	case StatePrice:
		return c.sendMessage(chatID, msgInvalidPrice)
	case StateStock:
		return c.sendMessage(chatID, msgInvalidStock)
	case StateImages:
		return c.sendMessage(chatID, fmt.Sprintf("❌ 最多只能上傳 5 張圖片，目前已上傳 %d 張", 5))
	default:
		return c.sendMessage(chatID, msgInvalidInput)
	}
}
