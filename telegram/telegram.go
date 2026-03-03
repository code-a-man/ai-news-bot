package telegram

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const maxMessageLen = 4096

type Client struct {
	bot *tgbotapi.BotAPI
}

func New(token string) (*Client, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram: %w", err)
	}
	return &Client{bot: bot}, nil
}

func (c *Client) SendToChats(chatIDs []string, text string) error {
	for _, chatID := range chatIDs {
		if err := c.Send(chatID, text); err != nil {
			return fmt.Errorf("send to %s: %w", chatID, err)
		}
	}
	return nil
}

func (c *Client) SendSingle(chatID string, text string) (int, error) {
	var msg tgbotapi.MessageConfig
	if id, err := strconv.ParseInt(chatID, 10, 64); err == nil {
		msg = tgbotapi.NewMessage(id, text)
	} else {
		msg = tgbotapi.NewMessageToChannel(chatID, text)
	}
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	m, err := c.bot.Send(msg)
	if err != nil {
		return 0, fmt.Errorf("send: %w", err)
	}
	return m.MessageID, nil
}

func (c *Client) EditMessage(chatID string, messageID int, text string) error {
	var cfg tgbotapi.EditMessageTextConfig
	if id, err := strconv.ParseInt(chatID, 10, 64); err == nil {
		cfg = tgbotapi.NewEditMessageText(id, messageID, text)
	} else {
		cfg = tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{ChannelUsername: chatID, MessageID: messageID},
			Text:     text,
		}
	}
	cfg.ParseMode = "Markdown"
	cfg.DisableWebPagePreview = true
	_, err := c.bot.Send(cfg)
	return err
}

func (c *Client) Send(chatID string, text string) error {
	chunks := splitMessage(text)
	for i, chunk := range chunks {
		var msg tgbotapi.MessageConfig
		if id, err := strconv.ParseInt(chatID, 10, 64); err == nil {
			msg = tgbotapi.NewMessage(id, chunk)
		} else {
			msg = tgbotapi.NewMessageToChannel(chatID, chunk)
		}
		msg.ParseMode = "Markdown"
		msg.DisableWebPagePreview = true
		if len(chunks) > 1 {
			msg.Text = chunk + "\n\n(" + fmt.Sprint(i+1) + "/" + fmt.Sprint(len(chunks)) + ")"
		}
		if _, err := c.bot.Send(msg); err != nil {
			return fmt.Errorf("send: %w", err)
		}
	}
	return nil
}

func splitMessage(text string) []string {
	if len(text) <= maxMessageLen {
		return []string{text}
	}
	var chunks []string
	for len(text) > 0 {
		end := maxMessageLen
		if end > len(text) {
			end = len(text)
		} else {
			idx := strings.LastIndex(text[:end], "\n\n")
			if idx > maxMessageLen/2 {
				end = idx + 2
			}
		}
		chunks = append(chunks, strings.TrimSpace(text[:end]))
		text = strings.TrimSpace(text[end:])
	}
	return chunks
}
