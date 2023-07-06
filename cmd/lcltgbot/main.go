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

	settings := GetSettings()
	textsettings := GetText()

	db := lcltgbot.NewSqliteDb(settings)

	api, err := tgbotapi.NewBotAPI(settings.Key)

	if err != nil {
		log.Fatal(err)
	}

	handl := handlers.NewHandlers(api, db, settings, textsettings)

	application := app.New(api, handl)

	application.Start()
}

func GetSettings() *models.AppSettings {
	botdatafile, err := os.Open("appsettings.json")

	if err != nil {
		log.Fatal(err)
	}

	settingsdecoder := json.NewDecoder(botdatafile)

	var botdata models.AppSettings

	if err = settingsdecoder.Decode(&botdata); err != nil {
		log.Fatal(err)
	}

	return &botdata
}

func GetText() *models.TextSettings {
	textsettingsfile, err := os.Open("assets/translations/ru.json")

	if err != nil {
		log.Fatal(err)
	}

	settingsdecoder := json.NewDecoder(textsettingsfile)

	var textSettings models.TextSettings

	if err = settingsdecoder.Decode(&textSettings); err != nil {
		log.Fatal(err)
	}

	return &textSettings
}
