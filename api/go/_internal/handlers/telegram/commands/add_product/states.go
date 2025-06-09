package add_product

import (
	"context"

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

// UI Message constants for FSM states
const (
	promptInit        = "歡迎使用商品上架功能！讓我們開始吧。"
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

type AddProductState interface {
	Name() string
	Buttons() []tgbotapi.InlineKeyboardButton
	Prompt() string
	Send(msg *tgbotapi.Message) error
	Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error
}

func AsAddProductState(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(AddProductState)),
		fx.ResultTags(`group:"add_product_states"`),
	)
}

// StateInit - Initial state, no buttons needed
type AddProductStateInit struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateInit(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateInit{
		botAPI: botAPI,
	}
}

func (s *AddProductStateInit) Name() string {
	return StateInit
}

func (s *AddProductStateInit) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{}
}

func (s *AddProductStateInit) Prompt() string {
	return promptInit
}

func (s *AddProductStateInit) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateInit) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateName - Required field, only cancel/pause options
type AddProductStateName struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateName(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateName{
		botAPI: botAPI,
	}
}

func (s *AddProductStateName) Name() string {
	return StateName
}

func (s *AddProductStateName) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateName) Prompt() string {
	return promptName
}

func (s *AddProductStateName) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateName) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateCategory - Required field, only cancel/pause options
type AddProductStateCategory struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateCategory(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateCategory{
		botAPI: botAPI,
	}
}

func (s *AddProductStateCategory) Name() string {
	return StateCategory
}

func (s *AddProductStateCategory) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateCategory) Prompt() string {
	return promptCategory
}

func (s *AddProductStateCategory) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateCategory) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StatePrice - Required field, only cancel/pause options
type AddProductStatePrice struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStatePrice(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStatePrice{
		botAPI: botAPI,
	}
}

func (s *AddProductStatePrice) Name() string {
	return StatePrice
}

func (s *AddProductStatePrice) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStatePrice) Prompt() string {
	return promptPrice
}

func (s *AddProductStatePrice) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStatePrice) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateStock - Required field, only cancel/pause options
type AddProductStateStock struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateStock(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateStock{
		botAPI: botAPI,
	}
}

func (s *AddProductStateStock) Name() string {
	return StateStock
}

func (s *AddProductStateStock) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateStock) Prompt() string {
	return promptStock
}

func (s *AddProductStateStock) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateStock) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateDescription - Optional field, can be skipped
type AddProductStateDescription struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateDescription(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateDescription{
		botAPI: botAPI,
	}
}

func (s *AddProductStateDescription) Name() string {
	return StateDescription
}

func (s *AddProductStateDescription) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", "skip_description"),
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateDescription) Prompt() string {
	return promptDescription
}

func (s *AddProductStateDescription) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateDescription) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateSpecs - Multi-input optional field, needs done/skip buttons
type AddProductStateSpecs struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateSpecs(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateSpecs{
		botAPI: botAPI,
	}
}

func (s *AddProductStateSpecs) Name() string {
	return StateSpecs
}

func (s *AddProductStateSpecs) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ 完成", "done_specs"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", "skip_specs"),
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateSpecs) Prompt() string {
	return promptSpecs
}

func (s *AddProductStateSpecs) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateSpecs) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateImages - Multi-input optional field, needs done/skip buttons
type AddProductStateImages struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateImages(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateImages{
		botAPI: botAPI,
	}
}

func (s *AddProductStateImages) Name() string {
	return StateImages
}

func (s *AddProductStateImages) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ 完成", "done_images"),
		tgbotapi.NewInlineKeyboardButtonData("⏭️ 跳過", "skip_images"),
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateImages) Prompt() string {
	return promptImages
}

func (s *AddProductStateImages) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	buttons := s.Buttons()
	if len(buttons) > 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(buttons...),
		)
		message.ReplyMarkup = keyboard
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateImages) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	return s.Send(fsmCtx.Message)
}

// StateConfirm - Final confirmation, only confirm/cancel
type AddProductStateConfirm struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateConfirm(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateConfirm{
		botAPI: botAPI,
	}
}

func (s *AddProductStateConfirm) Name() string {
	return StateConfirm
}

func (s *AddProductStateConfirm) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ 確認", "confirm"),
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
	}
}

func (s *AddProductStateConfirm) Prompt() string {
	return ""
}

func (s *AddProductStateConfirm) Send(msg *tgbotapi.Message) error {
	// Note: StateConfirm doesn't send its own message as it's handled by sendSummary
	return nil
}

func (s *AddProductStateConfirm) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	// TODO: Implement confirm logic - will handle summary display
	return nil
}

// StateCompleted - Final state, no buttons needed
type AddProductStateCompleted struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateCompleted(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateCompleted{
		botAPI: botAPI,
	}
}

func (s *AddProductStateCompleted) Name() string {
	return StateCompleted
}

func (s *AddProductStateCompleted) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{}
}

func (s *AddProductStateCompleted) Prompt() string {
	return msgSuccess
}

func (s *AddProductStateCompleted) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateCompleted) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	// TODO: Implement completion logic
	return nil
}

// StateCancelled - Final state, no buttons needed
type AddProductStateCancelled struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStateCancelled(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStateCancelled{
		botAPI: botAPI,
	}
}

func (s *AddProductStateCancelled) Name() string {
	return StateCancelled
}

func (s *AddProductStateCancelled) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{}
}

func (s *AddProductStateCancelled) Prompt() string {
	return msgCancelled
}

func (s *AddProductStateCancelled) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateCancelled) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	// TODO: Implement cancellation logic
	return nil
}

// StatePaused - Paused state, offer resume option
type AddProductStatePaused struct {
	botAPI *tgbotapi.BotAPI
}

func NewAddProductStatePaused(botAPI *tgbotapi.BotAPI) AddProductState {
	return &AddProductStatePaused{
		botAPI: botAPI,
	}
}

func (s *AddProductStatePaused) Name() string {
	return StatePaused
}

func (s *AddProductStatePaused) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("▶️ 繼續", "resume"),
		tgbotapi.NewInlineKeyboardButtonData("🔄 重新開始", "restart"),
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
	}
}

func (s *AddProductStatePaused) Prompt() string {
	return msgPaused
}

func (s *AddProductStatePaused) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	buttons := s.Buttons()
	if len(buttons) > 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(buttons...),
		)
		message.ReplyMarkup = keyboard
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStatePaused) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	// TODO: Implement pause logic
	return nil
}

// Factory function to create state instances based on state name
func NewAddProductStateMap(states []AddProductState) map[string]AddProductState {
	statesMap := make(map[string]AddProductState)
	for _, state := range states {
		statesMap[state.Name()] = state
	}
	return statesMap
}
