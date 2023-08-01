package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/otiai10/openaigo"
)

var history sync.Map

const maxHistoryPerUser = 5

func main() {

	client := openaigo.NewClient(os.Getenv("OPENAI_API_KEY"))
	// Replace with your own bot token
	botToken := os.Getenv("BOT_TOKEN")

	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	// Set up a handler function to process updates
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	// Process updates in a loop
	for update := range updates {
		istyping(bot, update.Message.Chat.ID)

		if update.Message.IsCommand() && update.Message.Command() == "reset" {
			history.Delete(update.Message.From.ID)
			if _, err = send(bot, update.Message.Chat.ID, "Done\\!", 0); err != nil {
				log.Println(err)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		// Get the user's input
		input := update.Message.Text

		res, err := getTextResponse(bot, update.Message.Chat.ID, update.Message.From.ID, client, input)
		if err != nil {
			res = err.Error()
			log.Println(err)
		}
		res = sanitize(res)
		// Send the input back to the user
		if _, err = send(bot, update.Message.Chat.ID, res, 0); err != nil {
			log.Println(err)
		}
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string, id int) (int, error) {
	if id == 0 {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = tgbotapi.ModeMarkdownV2
		m, err := bot.Send(msg)
		if err == nil {
			return m.MessageID, nil
		}
		return 0, err
	}

	edit := tgbotapi.NewEditMessageText(chatID, id, text)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	m, err := bot.Send(edit)
	if err == nil {
		return m.MessageID, nil
	}
	return 0, err
}

func getTextResponse(bot *tgbotapi.BotAPI, chatID int64, clientID int64, client *openaigo.Client, text string) (string, error) {
	ticker := time.NewTicker(time.Second * 2)
	chunks := ""
	id := 0
	var messages []openaigo.Message
	if messagesHistory, ok := history.Load(clientID); ok {
		if chatMessages, ok := messagesHistory.([]openaigo.Message); ok {
			maxHistory := len(chatMessages) - maxHistoryPerUser
			if len(chatMessages) < maxHistoryPerUser {
				maxHistory = 0
			}
			messages = append(messages, chatMessages[maxHistory:]...)
		}
	}
	messages = append(messages, openaigo.Message{
		Role:    "user",
		Content: text,
	})

	request := openaigo.ChatCompletionRequestBody{
		Model:       "gpt-4",
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.2,
		User:        strconv.FormatInt(clientID, 10),
		Stream:      true,
		StreamCallback: func(res openaigo.ChatCompletionResponse, done bool, err error) {

			if !done {
				completionchunch := res.Choices[0].Delta.Content

				completionchunch = sanitize(completionchunch)
				chunks += completionchunch

				select {
				case <-ticker.C:
					istyping(bot, chatID)
					newid, err := send(bot, chatID, chunks, id)
					if err != nil {
						log.Println(err)
					}
					id = newid
				default:
				}

			}
			if done {
				if _, err = send(bot, chatID, chunks, id); err != nil {
					log.Println(err)
				}
				if _, err = send(bot, chatID, "done", 0); err != nil {
					log.Println(err)
				}
			}

		},
	}

	_, err := client.ChatCompletion(nil, request)
	if err != nil {
		return "", err
	}

	//messages = append(messages, openaigo.Message{
	//	Role:    "system",
	//	Content: response.Choices[0].Message.Content,
	//})

	history.Store(clientID, messages)

	return "let me think...", nil
}

func sanitize(res string) string {
	res = strings.ReplaceAll(res, "_", "\\_")
	res = strings.ReplaceAll(res, "*", "\\*")
	res = strings.ReplaceAll(res, "[", "\\[")
	res = strings.ReplaceAll(res, "]", "\\]")
	res = strings.ReplaceAll(res, "`", "\\`")
	res = strings.ReplaceAll(res, "(", "\\(")
	res = strings.ReplaceAll(res, ")", "\\)")
	res = strings.ReplaceAll(res, "~", "\\~")
	res = strings.ReplaceAll(res, "#", "\\#")
	res = strings.ReplaceAll(res, "+", "\\+")
	res = strings.ReplaceAll(res, "-", "\\-")
	res = strings.ReplaceAll(res, "=", "\\=")
	res = strings.ReplaceAll(res, "|", "\\|")
	res = strings.ReplaceAll(res, "{", "\\{")
	res = strings.ReplaceAll(res, "}", "\\}")
	res = strings.ReplaceAll(res, ".", "\\.")
	res = strings.ReplaceAll(res, "!", "\\!")
	res = strings.ReplaceAll(res, ">", "\\>")
	res = strings.ReplaceAll(res, "$", "\\$")
	res = strings.ReplaceAll(res, "<", "\\<")

	return res
}

func istyping(bot *tgbotapi.BotAPI, chatID int64) {
	_, err := bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
	if err != nil {
		log.Println(err)
	}
}
