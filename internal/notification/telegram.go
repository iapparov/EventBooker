package notification

import (
	"fmt"
	"log"
	"strconv"

	"eventbooker/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	wbzlog "github.com/wb-go/wbf/zlog"
)

// TelegramSender sends cancellation notifications via Telegram.
type TelegramSender struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramSender creates a new TelegramSender.
func NewTelegramSender(cfg *config.AppConfig) *TelegramSender {
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to create Telegram bot")
		return nil
	}

	ts := &TelegramSender{bot: bot}
	go ts.listenForStartCommand()
	return ts
}

// Send sends a booking cancellation message via Telegram.
func (t *TelegramSender) Send(tg, eventName string, persons int) error {
	chatID, err := strconv.Atoi(tg)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	msg := tgbotapi.NewMessage(int64(chatID),
		fmt.Sprintf("Your booking on %d persons on event: %s just cancelled", persons, eventName))
	_, err = t.bot.Send(msg)
	return err
}

func (t *TelegramSender) listenForStartCommand() {
	log.Println("Telegram listener started...")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			chatID := update.Message.Chat.ID
			username := update.Message.From.UserName

			log.Printf("[TG] User %s started bot, chat_id=%d", username, chatID)

			msg := tgbotapi.NewMessage(chatID,
				fmt.Sprintf("👋 Привет, %s!\n\nТвой chat_id: `%d`\nОтправь его в приложение, чтобы получать уведомления.",
					username, chatID))

			if _, err := t.bot.Send(msg); err != nil {
				wbzlog.Logger.Error().Err(err).Msg("failed to send Telegram message")
			}
		}
	}
}
