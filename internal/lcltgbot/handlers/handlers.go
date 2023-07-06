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

const DEBUG = true

type Database interface {
	Register(chatid int64) (*models.User, error)
	GetUser(chatid int64) (*models.User, error)
	ChangeUserState(user *models.User, state models.BotState) (*models.User, error)
	ChangeAdTitle(user *models.User, title string) (*models.User, error)
	ChangeAdDescription(user *models.User, descr string) (*models.User, error)
	ChangeAdPrice(user *models.User, price float64) (*models.User, error)
	ChangeAdCity(user *models.User, city string) (*models.User, error)
	ChangeAdEditing(user *models.User, editing bool) (*models.User, error)
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
	default:
		if err := h.SendMessage(user, "Команда или текст не распознаны и/или не подходят в этом контексте!"); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) CheckIfFlowMessageIsValid(user *models.User, message *tgbotapi.Message) error {
	if message.IsCommand() {
		if message.Text != commands.CancelFlow {
			h.SendMessage(user, fmt.Sprintf("Сейчас вы находитесь в \"цепи набора\". Использовать команды нельзя. Пройдите всю цепь или используйте %s, чтобы отменить цепь!", commands.CancelFlow))
			return nil
		}

		if _, err := h.db.ChangeUserState(user, models.StateNONE); err != nil {
			return err
		}

		if err := h.SendMessage(user, "Набор успешно отменен!"); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) HandleCommandFlow(user *models.User, message *tgbotapi.Message) error {
	chatid := user.Chatid

	username := message.From.UserName

	if err := h.CheckIfFlowMessageIsValid(user, message); err != nil {
		return err
	}

	switch user.Context.State {
	case models.StateWaitingForCTitle:
		user, err := h.db.ChangeAdTitle(user, message.Text)
		if err != nil {
			return err
		}

		if _, err := h.GoNextIfCreatingElseDropEditing(
			user,
			models.StateWaitingForCDescription,
			chatid,
			"Введите описание товара.\n\n\nПрим. цена и город будут указываться далее, писать их в описании нет необходимости",
			username,
		); err != nil {
			return err
		}

	case models.StateWaitingForCDescription:
		user, err := h.db.ChangeAdDescription(user, message.Text)

		if err != nil {
			return err
		}

		if _, err := h.GoNextIfCreatingElseDropEditing(
			user,
			models.StateWaitingForCPrice,
			chatid,
			"Введите цену товара (руб).\n\n\nПрим. обязательно число!",
			username,
		); err != nil {
			return err
		}

	case models.StateWaitingForCPrice:
		price, _ := strconv.Atoi(message.Text)

		user, err := h.db.ChangeAdPrice(user, float64(price))

		if err != nil {
			return err
		}

		if _, err := h.GoNextIfCreatingElseDropEditing(
			user,
			models.StateWaitingForCCity,
			chatid,
			"Введите город.",
			username,
		); err != nil {
			return err
		}

	case models.StateWaitingForCCity:
		user, err := h.db.ChangeAdCity(user, message.Text)

		if err != nil {
			return err
		}

		if !user.Context.Advertisement.Editing {
			if err := h.SendPreview(user, message.From.UserName); err != nil {
				return err
			}
		}

		if _, err := h.GoNextIfCreatingElseDropEditing(
			user,
			models.StateNONE,
			chatid,
			"Готово! Так будет выглядеть ваще объявление!",
			username,
		); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) GoNextIfCreatingElseDropEditing(user *models.User, state models.BotState, chatid int64, messagetext string, username string) (*models.User, error) {
	if !user.Context.Advertisement.Editing {
		user, err := h.db.ChangeUserState(user, state)

		if err != nil {
			return nil, err
		}

		if err := h.SendMessage(user, messagetext); err != nil {
			return nil, err
		}

		return user, nil
	}

	return h.AfterEdited(user, username)
}

func (h *Handlers) AfterEdited(user *models.User, username string) (*models.User, error) {
	_, err := h.bot.Send(h.CreateNewAdMessage(user, username, tgbotapi.ModeHTML))
	if err != nil {
		return nil, err
	}

	err = h.DropEditing(user)
	if err != nil {
		return nil, err
	}
	user, err = h.DropUserState(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *Handlers) DropEditing(user *models.User) error {
	if _, err := h.db.ChangeAdEditing(user, false); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) DropUserState(user *models.User) (*models.User, error) {
	user, err := h.db.ChangeUserState(user, models.StateNONE)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (h *Handlers) CreateNewAdMessage(user *models.User, username string, parsemode string) tgbotapi.MessageConfig {
	if DEBUG {
		username = "<b>недоступно</b>"
	}

	message := tgbotapi.NewMessage(user.Chatid, formatters.FormatAdToMessageString(user.Context.Advertisement, username))
	message.ParseMode = parsemode
	message.ReplyMarkup = h.GetPreviewMarkup(username)

	return message
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

func (h *Handlers) SendPreview(user *models.User, username string) error {
	if _, err := h.bot.Send(h.CreateNewAdMessage(user, username, tgbotapi.ModeHTML)); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) GetPreviewMarkup(username string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(commands.SendButtonPair.ParamName, fmt.Sprintf("%s:%s", commands.SendButtonPair.ParamValue, username)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeTitleButton.ParamName, fmt.Sprintf("%s:%s:%d", commands.ChangeTitleButton.ParamValue, username, models.StateWaitingForCTitle)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeDescriptionButton.ParamName, fmt.Sprintf("%s:%s:%d", commands.ChangeDescriptionButton.ParamValue, username, models.StateWaitingForCDescription)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangePriceButton.ParamName, fmt.Sprintf("%s:%s:%d", commands.ChangePriceButton.ParamValue, username, models.StateWaitingForCPrice)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeCityButton.ParamName, fmt.Sprintf("%s:%s:%d", commands.ChangeCityButton.ParamValue, username, models.StateWaitingForCCity)),
		),
	)
}

func (h *Handlers) HandleCallbackQuery(query *tgbotapi.CallbackQuery) error {
	querydata := strings.Split(query.Data, ":")

	if len(querydata) < 2 {
		return errors.New("to low parameters for callback query")
	}

	user, err := h.db.GetUser(query.Message.Chat.ID)

	username := querydata[1]

	if err != nil {
		return err
	}

	switch querydata[0] {
	case commands.SendButtonPair.ParamValue:
		message := tgbotapi.NewMessageToChannel("@lcltg", formatters.FormatAdToMessageString(user.Context.Advertisement, username))
		message.ParseMode = tgbotapi.ModeHTML

		if DEBUG {
			message.DisableNotification = true
		}

		_, err = h.bot.Send(message)
		if err != nil {
			return err
		}

		query.Message.ReplyMarkup = nil
	case commands.ChangeValueCommandData:
		statenum, err := strconv.Atoi(querydata[2])

		if err != nil {
			return err
		}

		state := models.BotState(statenum)

		if _, err := h.db.ChangeAdEditing(user, true); err != nil {
			return err
		}

		user, err = h.db.ChangeUserState(user, state)
		if err != nil {
			return err
		}

		if err = h.SendMessage(user, "Введите новое значение параметра"); err != nil {
			return err
		}
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

	if err := h.SendMessage(models.NewUser(chatid, nil), "Доступ к боту разрешен только по ключу. Введите ключ!"); err != nil {
		return nil, err
	}

	return nil, nil
}
