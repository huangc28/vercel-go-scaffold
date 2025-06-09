package telegram

import (
	"context"
	"fmt"
	"log"

	"github/huangc28/kikichoice-be/api/go/_internal/handlers/telegram/commands"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/fx"
)

type ReplyProcessor struct {
	commandDAO      *commands.CommandDAO
	commandHandlers map[commands.BotCommand]commands.CommandHandler
}

type ReplyProcessorParams struct {
	fx.In

	CommandDAO      *commands.CommandDAO
	CommandHandlers map[commands.BotCommand]commands.CommandHandler
}

func NewReplyProcessor(p ReplyProcessorParams) *ReplyProcessor {
	return &ReplyProcessor{
		commandDAO:      p.CommandDAO,
		commandHandlers: p.CommandHandlers,
	}
}

func (r *ReplyProcessor) Process(ctx context.Context, reply *tgbotapi.Message) error {
	session, err := r.commandDAO.GetUserSession(ctx, reply.From.ID, reply.Chat.ID, reply.Command())
	if err != nil {
		return fmt.Errorf("failed to get user session: %w", err)
	}

	// Determine what command handler to use.
	log.Printf("** 1 %v", session.ExpectedReplyMessageID)

	// Is the incoming message a reply to the exepcted reply message?
	if reply.ReplyToMessage.MessageID == int(session.ExpectedReplyMessageID.Int64) {
		r.commandHandlers[commands.BotCommand(session.SessionType)].Reply(ctx, reply)
	}

	return nil
}
