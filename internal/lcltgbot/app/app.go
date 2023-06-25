package app

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
	"log"
	"os"
)

const (
	StartCommand = "/start"
)

type Database interface {
	Register(chatid int64) (*models.User, error)
	GetUser(chatid int64) (*models.User, error)
}

type Handlers interface {
	HandleSingleCommand(user *models.User) error
	HandleCommandFlow(user *models.User) error
}

type App struct {
	botapi   *tgbotapi.BotAPI
	db       Database
	handlers Handlers
}

func New(botdatapath string, db Database, handlers Handlers) *App {
	botdatafile, err := os.Open("botdata.json")

	if err != nil {
		log.Fatal(err)
	}

	jsondecoder := json.NewDecoder(botdatafile)

	var botdata models.BotData

	if err = jsondecoder.Decode(&botdata); err != nil {
		log.Fatal(err)
	}

	botapi, err := tgbotapi.NewBotAPI(botdata.Key)

	if err != nil {
		log.Fatal(err)
	}

	return &App{
		botapi:   botapi,
		db:       db,
		handlers: handlers,
	}
}

func (a *App) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.botapi.GetUpdatesChan(u)

	for update := range updates {
		a.HandleUpdate(update)
	}
}

func (a *App) HandleUpdate(update tgbotapi.Update) {
	if update.Message.Text == StartCommand {
		user, _ := a.db.Register(update.Message.Chat.ID)
		if err := a.handlers.HandleSingleCommand(user); err != nil {
			log.Fatal(err)
		}
	}
}
