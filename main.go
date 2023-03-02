package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/otiai10/openaigo"
)

func main() {
	client := openaigo.NewClient(os.Getenv("OPENAI_API_KEY"))
	client.BaseURL = "https://api.openai.com/v1/chat"
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
		if update.Message == nil {
			continue
		}

		// Get the user's input
		input := update.Message.Text

		res, err := getResponse(update.Message.From.ID, client, input)
		if err != nil {
			res = err.Error()
			log.Println(err)
		}
		// Send the input back to the user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, res)
		_, err = bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func getResponse(clientID int, client *openaigo.Client, text string) (string, error) {
	request := openaigo.ChatCompletionRequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []openaigo.ChatMessage{
			{Role: "user", Content: "Response in Telegram format with more details unless I tell you to do differently."},
			{Role: "user", Content: text},
		},
		MaxTokens:   1000,
		Temperature: 0.2,
		User:        strconv.Itoa(clientID),
	}
	response, err := client.Chat(nil, request)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response.Choices[0].Message.Content), nil
}
