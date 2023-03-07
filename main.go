package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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
	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Process updates in a loop
	for update := range updates {
		if update.Message.IsCommand() && update.Message.Command() == "reset" {
			history.Delete(update.Message.From.ID)
			if err = send(bot, update.Message.Chat.ID, "Done!"); err != nil {
				log.Println(err)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		// Get the user's input
		input := update.Message.Text

		res, err := getTextResponse(update.Message.From.ID, client, input)
		if err != nil {
			res = err.Error()
			log.Println(err)
		}
		// Send the input back to the user
		if err = send(bot, update.Message.Chat.ID, res); err != nil {
			log.Println(err)
		}
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}

func getTextResponse(clientID int, client *openaigo.Client, text string) (string, error) {
	var messages []openaigo.ChatMessage
	if messagesHistory, ok := history.Load(clientID); ok {
		if chatMessages, ok := messagesHistory.([]openaigo.ChatMessage); ok {
			maxHistory := len(chatMessages) - maxHistoryPerUser
			if len(chatMessages) < maxHistoryPerUser {
				maxHistory = 0
			}
			messages = append(messages, chatMessages[maxHistory:]...)
		}
	}
	messages = append(messages, openaigo.ChatMessage{
		Role:    "user",
		Content: text,
	})

	request := openaigo.ChatCompletionRequestBody{
		Model:       "gpt-3.5-turbo",
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.2,
		User:        strconv.Itoa(clientID),
	}

	response, err := client.Chat(nil, request)
	if err != nil {
		return "", err
	}

	messages = append(messages, openaigo.ChatMessage{
		Role:    "system",
		Content: response.Choices[0].Message.Content,
	})

	history.Store(clientID, messages)

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
