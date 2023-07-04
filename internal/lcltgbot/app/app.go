package app

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
	"log"
)

type Handlers interface {
	HandleSingleCommand(user *models.User, message *tgbotapi.Message) error
	HandleCommandFlow(user *models.User, message *tgbotapi.Message) error
	HandleMessage(message *tgbotapi.Message) error
	HandleCallbackQuery(query *tgbotapi.CallbackQuery) error
}

type App struct {
	botapi   *tgbotapi.BotAPI
	handlers Handlers
}

func New(botapi *tgbotapi.BotAPI, handlers Handlers) *App {
	return &App{
		botapi:   botapi,
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
	if update.Message != nil {
		if err := a.handlers.HandleMessage(update.Message); err != nil {
			log.Fatal(err)
		}
	} else if update.CallbackQuery != nil {
		a.handlers.HandleCallbackQuery(update.CallbackQuery)
	}
}
