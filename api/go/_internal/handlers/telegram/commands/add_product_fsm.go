package commands

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
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
	UserID           int64
	ChatID           int64
	Message          *tgbotapi.Message
	UserState        *UserState
	AddProductStates map[string]AddProductState
	Command          *AddProductCommand
}

// NewAddProductFSM creates a new FSM instance with all events and callbacks
func NewAddProductFSM(
	c *AddProductCommand,
	userID, chatID int64,
	state *UserState,
	msg *tgbotapi.Message,
	addProductStates map[string]AddProductState,
) *fsm.FSM {
	fsmCtx := &FSMContext{
		UserID:           userID,
		ChatID:           chatID,
		Message:          msg,
		Command:          c,
		UserState:        state,
		AddProductStates: addProductStates,
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
			"enter_" + StateSKU: func(ctx context.Context, e *fsm.Event) {
				fsmCtx.AddProductStates[StateSKU].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateName: func(ctx context.Context, e *fsm.Event) {
				fsmCtx.AddProductStates[StateName].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateCategory: func(ctx context.Context, e *fsm.Event) {
				fsmCtx.AddProductStates[StateCategory].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StatePrice: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StatePrice)
				fsmCtx.AddProductStates[StatePrice].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateStock: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StateStock)
				fsmCtx.AddProductStates[StateStock].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateDescription: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StateDescription)
				fsmCtx.AddProductStates[StateDescription].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateSpecs: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StateSpecs)
				fsmCtx.AddProductStates[StateSpecs].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateImages: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StateImages)
				fsmCtx.AddProductStates[StateImages].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateConfirm: func(ctx context.Context, e *fsm.Event) {
				log.Println("enter_" + StateConfirm)
				fsmCtx.AddProductStates[StateConfirm].Enter(ctx, e, fsmCtx)
			},
			"enter_" + StateCompleted: func(ctx context.Context, e *fsm.Event) {},
			"enter_" + StateCancelled: func(ctx context.Context, e *fsm.Event) {},
			"enter_" + StatePaused:    func(ctx context.Context, e *fsm.Event) {},

			// Before event callbacks (validation)
			"before_" + EventNext: func(ctx context.Context, e *fsm.Event) { c.validateInput(ctx, e, fsmCtx) },

			// After event callbacks (data storage)
			"after_" + EventNext: func(ctx context.Context, e *fsm.Event) { c.storeInput(ctx, e, fsmCtx) },
		},
	)
}

// FSM Event Callbacks

// validateInput checks if input is valid before proceeding with FSM event
func (c *AddProductCommand) validateInput(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	switch e.Src {
	case StatePrice:
		if !c.isValidPrice(fsmCtx.UserState.CurrentInput) {
			e.Cancel(fmt.Errorf("invalid price format"))
			return
		}
	case StateStock:
		if !c.isValidStock(fsmCtx.UserState.CurrentInput) {
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
	return !(fsmCtx.Message != nil && fsmCtx.Message.Photo != nil && len(fsmCtx.UserState.ImageFileIDs) >= 5)
}

func (c *AddProductCommand) storeInput(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) {
	log.Println("storeInput!")
	// switch e.Src {
	// case StateSKU:
	// 	fsmCtx.UserState.Product.SKU = fsmCtx.UserState.CurrentInput
	// case StateName:
	// 	fsmCtx.UserState.Product.Name = fsmCtx.UserState.CurrentInput
	// case StateCategory:
	// 	fsmCtx.UserState.Product.Category = fsmCtx.UserState.CurrentInput
	// case StatePrice:
	// 	price, _ := strconv.ParseFloat(fsmCtx.UserState.CurrentInput, 64)
	// 	fsmCtx.UserState.Product.Price = price
	// case StateStock:
	// 	stock, _ := strconv.Atoi(fsmCtx.UserState.CurrentInput)
	// 	fsmCtx.UserState.Product.Stock = stock
	// case StateDescription:
	// 	fsmCtx.UserState.Product.Description = fsmCtx.UserState.CurrentInput
	// case StateSpecs:
	// 	if fsmCtx.State.CurrentInput != "/done" {
	// 		fsmCtx.State.Specs = append(fsmCtx.State.Specs, fsmCtx.State.CurrentInput)
	// 		// Send feedback for specs
	// 		c.sendMessage(fsmCtx.ChatID, msgSpecAdded)
	// 	}
	// case StateImages:
	// 	if fsmCtx.Message != nil && fsmCtx.Message.Photo != nil {
	// 		fileID := fsmCtx.Message.Photo[len(fsmCtx.Message.Photo)-1].FileID
	// 		fsmCtx.State.ImageFileIDs = append(fsmCtx.State.ImageFileIDs, fileID)

	// 		// Send feedback for images
	// 		const maxImages = 5
	// 		remaining := maxImages - len(fsmCtx.State.ImageFileIDs)
	// 		if remaining > 0 {
	// 			c.sendMessage(fsmCtx.ChatID, fmt.Sprintf(msgImageUploaded, len(fsmCtx.State.ImageFileIDs), maxImages, remaining))
	// 		} else {
	// 			c.sendMessage(fsmCtx.ChatID, fmt.Sprintf(msgImageLimitReached, len(fsmCtx.State.ImageFileIDs), maxImages))
	// 		}
	// 	}
	// }
}

// handleInvalidInput handles invalid input for current state
// func (c *AddProductCommand) handleInvalidInput(chatID int64, currentState, input string) error {
// 	switch currentState {
// 	case StatePrice:
// 		return c.sendMessage(chatID, msgInvalidPrice)
// 	case StateStock:
// 		return c.sendMessage(chatID, msgInvalidStock)
// 	case StateImages:
// 		return c.sendMessage(chatID, fmt.Sprintf("❌ 最多只能上傳 5 張圖片，目前已上傳 %d 張", 5))
// 	default:
// 		return c.sendMessage(chatID, msgInvalidInput)
// 	}
// }
