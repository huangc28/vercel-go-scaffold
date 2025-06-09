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
	log.Printf("* 2 %v", msg.Chat.ID)
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
	log.Printf(
		"** 1 %v %v %v",
		fsmCtx.Message.Chat.ID,
		fsmCtx.Message.From.ID,
		fsmCtx.UserState.FSMState,
	)
	fsmCtx.UserState.ExpectedReplyMessageID = &fsmCtx.Message.MessageID

	if err := s.commandDAO.UpsertUserSession(
		ctx,
		fsmCtx.Message.Chat.ID,
		fsmCtx.Message.From.ID,
		"add_product",
		fsmCtx.UserState,
	); err != nil {
		log.Printf("** 2 %v", err)
		return fmt.Errorf("failed to update user session in sku state enter: %w", err)
	}

	log.Printf("** 3 %v", fsmCtx.Message.Chat.ID)

	if err := s.Send(fsmCtx.Message); err != nil {
		return fmt.Errorf("failed to send message in sku state enter: %w", err)
	}

	return nil
}

var _ AddProductState = (*AddProductStateSKU)(nil)
