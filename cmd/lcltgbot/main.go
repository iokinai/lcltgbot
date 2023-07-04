package main

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/app"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/handlers"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
	"github.com/iokinai/lcltgbot/pkg/lcltgbot"
	"log"
	"os"
)

func main() {
	db := lcltgbot.NewSqliteDb()

	botdatafile, err := os.Open("botdata.json")

	if err != nil {
		log.Fatal(err)
	}

	jsondecoder := json.NewDecoder(botdatafile)

	var botdata models.BotData

	if err = jsondecoder.Decode(&botdata); err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	api, err := tgbotapi.NewBotAPI(botdata.Key)

	if err != nil {
		log.Fatal(err)
	}

	handl := handlers.NewHandlers(api, db, botdata.SecretKey)

	application := app.New(api, handl)

	application.Start()
}
