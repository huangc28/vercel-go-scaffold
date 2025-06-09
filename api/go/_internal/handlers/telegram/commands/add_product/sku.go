package add_product

import (
	"context"
	"fmt"
	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/looplab/fsm"
	"go.uber.org/fx"
)

type AddProductStateSKU struct {
	botAPI     *tgbotapi.BotAPI
	commandDAO *commands.CommandDAO
}

type AddProductStateSKUParams struct {
	fx.In

	BotAPI     *tgbotapi.BotAPI
	CommandDAO *commands.CommandDAO
}

func NewAddProductStateSKU(p AddProductStateSKUParams) AddProductState {
	return &AddProductStateSKU{
		botAPI:     p.BotAPI,
		commandDAO: p.CommandDAO,
	}
}

func (s *AddProductStateSKU) Name() string {
	return StateSKU
}

func (s *AddProductStateSKU) Buttons() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", "cancel"),
		tgbotapi.NewInlineKeyboardButtonData("💾 暫存", "pause"),
	}
}

func (s *AddProductStateSKU) Prompt() string {
	return promptSKU
}

func (s *AddProductStateSKU) Send(msg *tgbotapi.Message) error {
	message := tgbotapi.NewMessage(msg.Chat.ID, s.Prompt())
	message.ReplyToMessageID = msg.MessageID
	message.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply: true,
		Selective:  true,
	}
	_, err := s.botAPI.Send(message)
	return err
}

func (s *AddProductStateSKU) Enter(ctx context.Context, e *fsm.Event, fsmCtx *FSMContext) error {
	if err := s.commandDAO.UpdateExpectedReplyMessageID(
		ctx,
		fsmCtx.Message.Chat.ID,
		fsmCtx.Message.From.ID,
		fsmCtx.Command.Command().String(),
		fsmCtx.Message.MessageID,
	); err != nil {
		return fmt.Errorf("failed to update user session in sku state enter: %w", err)
	}

	if err := s.Send(fsmCtx.Message); err != nil {
		return fmt.Errorf("failed to send message in sku state enter: %w", err)
	}

	return nil
}

func (s *AddProductStateSKU) Reply(ctx context.Context, msg *tgbotapi.Message, fsm *fsm.FSM) error {
	log.Printf("** 3 %v", msg.ReplyToMessage.Text)
	log.Printf("** 3 %v", fsm.Current())
	return nil
}

var _ AddProductState = (*AddProductStateSKU)(nil)
