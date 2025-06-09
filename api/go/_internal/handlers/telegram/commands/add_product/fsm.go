package add_product

import (
	"context"
	"log"

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
	UserState        *AddProductSessionState
	AddProductStates map[string]AddProductState
	Command          *AddProductCommand
}

// NewAddProductFSM creates a new FSM instance with all events and callbacks
func NewAddProductFSM(
	c *AddProductCommand,
	userID, chatID int64,
	state *AddProductSessionState,
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
			"enter_" + StateInit: func(ctx context.Context, e *fsm.Event) {
				log.Printf("trigger enter_%s", StateInit)
				fsmCtx.AddProductStates[StateInit].Enter(ctx, e, fsmCtx)
			},
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
			"before_" + EventNext: func(ctx context.Context, e *fsm.Event) {},

			// After event callbacks (data storage)
			"after_" + EventNext: func(ctx context.Context, e *fsm.Event) {},
		},
	)
}
