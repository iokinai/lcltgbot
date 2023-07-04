package handlers

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/commands"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/formatters"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
	"log"
	"strconv"
	"strings"
)

type Database interface {
	Register(chatid int64) (*models.User, error)
	GetUser(chatid int64) (*models.User, error)
	ChangeUserState(user *models.User, state models.BotState) (*models.User, error)
	ChangeAdTitle(user *models.User, title string) error
	ChangeAdDescription(user *models.User, descr string) error
	ChangeAdPrice(user *models.User, price float64) error
	ChangeAdCity(user *models.User, city string) error
}

type Handlers struct {
	bot  *tgbotapi.BotAPI
	db   Database
	skey string
}

func NewHandlers(bot *tgbotapi.BotAPI, db Database, skey string) *Handlers {
	return &Handlers{bot: bot, db: db, skey: skey}
}

func (h *Handlers) HandleMessage(message *tgbotapi.Message) error {
	chatid := message.Chat.ID

	user, err := h.db.GetUser(chatid)

	if err != nil {
		user, err = h.AskForKey(chatid, message)

		if err != nil {
			return err
		}

		if user == nil {
			return nil
		}
	}

	if !user.Context.IsInFlow {
		if err := h.HandleSingleCommand(user, message); err != nil {
			return err
		}
	} else if err := h.HandleCommandFlow(user, message); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) HandleSingleCommand(user *models.User, message *tgbotapi.Message) error {
	switch message.Text {
	case commands.StartCommand:
		if err := h.HandleStart(user); err != nil {
			return err
		}
	case commands.AddAdCommand:
		if err := h.HandleAddAd(user); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) HandleCommandFlow(user *models.User, message *tgbotapi.Message) error {
	chatid := user.Chatid

	if message.IsCommand() {
		if message.Text != commands.CancelFlow {
			cantuse := tgbotapi.NewMessage(user.Chatid, fmt.Sprintf("Сейчас вы находитесь в \"цепи набора\". Использовать команды нельзя. Пройдите всю цепь или используйте %s, чтобы отменить цепь!", commands.CancelFlow))
			h.bot.Send(cantuse)
			return nil
		}

		if _, err := h.db.ChangeUserState(user, models.StateNONE); err != nil {
			return err
		}

		canceledmessage := tgbotapi.NewMessage(chatid, "Набор успешно отменен!")

		if _, err := h.bot.Send(canceledmessage); err != nil {
			return err
		}

		return nil
	}

	switch user.Context.State {
	case models.StateWaitingForCTitle:
		if err := h.db.ChangeAdTitle(user, message.Text); err != nil {
			return err
		}

		if _, err := h.db.ChangeUserState(user, models.StateWaitingForCDescription); err != nil {
			return err
		}

		nextmessage := tgbotapi.NewMessage(chatid, "Введите описание товара.\n\n\nПрим. цена и город будут указываться далее, писать их в описании нет необходимости")

		if _, err := h.bot.Send(nextmessage); err != nil {
			return err
		}

	case models.StateWaitingForCDescription:
		if err := h.db.ChangeAdDescription(user, message.Text); err != nil {
			return err
		}

		if _, err := h.db.ChangeUserState(user, models.StateWaitingForCPrice); err != nil {
			return err
		}

		nextmessage := tgbotapi.NewMessage(chatid, "Введите цену товара (руб).\n\n\nПрим. обязательно число!")

		if _, err := h.bot.Send(nextmessage); err != nil {
			return err
		}
	case models.StateWaitingForCPrice:
		price, _ := strconv.Atoi(message.Text)

		if err := h.db.ChangeAdPrice(user, float64(price)); err != nil {
			return err
		}

		if _, err := h.db.ChangeUserState(user, models.StateWaitingForCCity); err != nil {
			return err
		}

		nextmessage := tgbotapi.NewMessage(chatid, "Введите город.")

		if _, err := h.bot.Send(nextmessage); err != nil {
			return err
		}
	case models.StateWaitingForCCity:
		if err := h.db.ChangeAdCity(user, message.Text); err != nil {
			return err
		}

		if _, err := h.db.ChangeUserState(user, models.StateNONE); err != nil {
			return err
		}

		updateduser, err := h.db.GetUser(user.Chatid)

		if err != nil {
			return err
		}

		if err := h.FinalizeCreating(updateduser, message.From.UserName); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) HandleStart(user *models.User) error {
	if err := h.SendMessage(user, "STARTED [TEST]\n/add_ad - добавить объявление"); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) HandleAddAd(user *models.User) error {
	if err := h.SendMessage(user, "Отлично!\nПроцесс создания объявления разбит на несколько частей:\nУстановка названия\nУстановка описания\nУстановка цены\nУстановка города\n\nВведите название:"); err != nil {
		return err
	}

	if _, err := h.db.ChangeUserState(user, models.StateWaitingForCTitle); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) SendMessage(user *models.User, text string) error {
	message := tgbotapi.NewMessage(user.Chatid, text)
	_, err := h.bot.Send(message)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handlers) FinalizeCreating(user *models.User, username string) error {
	readymessage := tgbotapi.NewMessage(user.Chatid, "Готово! Так будет выглядеть ваше объявление:")

	if _, err := h.bot.Send(readymessage); err != nil {
		return err
	}

	preview := tgbotapi.NewMessage(user.Chatid, formatters.FormatAdToMessageString(user.Context.Advertisement, username))

	preview.ParseMode = tgbotapi.ModeHTML
	preview.ReplyMarkup = h.GetPreviewMarkup(username)

	if _, err := h.bot.Send(preview); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) GetPreviewMarkup(username string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(commands.SendButton, fmt.Sprintf("%s:%s", commands.SendButton, username))),
	)
}

func (h *Handlers) HandleCallbackQuery(query *tgbotapi.CallbackQuery) error {
	querydata := strings.Split(query.Data, ":")

	if len(querydata) < 2 {
		return errors.New("to low parameters for callback query")
	}

	switch querydata[0] {
	case commands.SendButton:
		user, err := h.db.GetUser(query.Message.Chat.ID)

		if err != nil {
			return err
		}

		message := tgbotapi.NewMessageToChannel("@lcltg", formatters.FormatAdToMessageString(user.Context.Advertisement, querydata[1]))
		message.ParseMode = tgbotapi.ModeHTML
		_, err = h.bot.Send(message)
		if err != nil {
			return err
		}

		query.Message.ReplyMarkup = nil
	}

	return nil
}

func (h *Handlers) AskForKey(chatid int64, message *tgbotapi.Message) (*models.User, error) {
	if message.Text == h.skey {
		user, err := h.db.Register(message.Chat.ID)
		if err != nil {
			log.Fatal(err)
		}

		if err := h.HandleStart(user); err != nil {
			return nil, err
		}

		return user, nil
	}

	askmessage := tgbotapi.NewMessage(chatid, "Доступ к боту разрешен только по ключу. Введите ключ!")

	if _, err := h.bot.Send(askmessage); err != nil {
		return nil, err
	}

	return nil, nil
}
