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
	Register(chatid int64, username string) (*models.User, error)
	GetUser(chatid int64) (*models.User, error)
	ChangeUserState(user *models.User, state models.BotState) (*models.User, error)
	ChangeAdTitle(user *models.User, title string) (*models.User, error)
	ChangeAdDescription(user *models.User, descr string) (*models.User, error)
	ChangeAdPrice(user *models.User, price float64) (*models.User, error)
	ChangeAdCity(user *models.User, city string) (*models.User, error)
	ChangeAdEditing(user *models.User, editing bool) (*models.User, error)
}

type Handlers struct {
	bot      *tgbotapi.BotAPI
	db       Database
	settings *models.AppSettings
	text     *models.TextSettings
}

func NewHandlers(bot *tgbotapi.BotAPI, db Database, settings *models.AppSettings, text *models.TextSettings) *Handlers {
	return &Handlers{bot: bot, db: db, settings: settings, text: text}
}

func (h *Handlers) HandleMessage(message *tgbotapi.Message) error {
	chatid := message.Chat.ID

	user, err := h.db.GetUser(chatid)

	if err != nil {
		user, err = h.AskForKey(message)

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
		if err := h.SendMessage(user, h.text.WrongCommand); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) CheckIfFlowMessageIsValid(user *models.User, message *tgbotapi.Message) error {
	if message.IsCommand() {
		if message.Text != commands.CancelFlow {
			h.SendMessage(user, fmt.Sprintf(h.text.InChainError, commands.CancelFlow))
			return errors.New("wrong command")
		}

		if _, err := h.db.ChangeUserState(user, models.StateNONE); err != nil {
			return err
		}

		if err := h.SendMessage(user, h.text.ChainCanceled); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) HandleCommandFlow(user *models.User, message *tgbotapi.Message) error {
	if err := h.CheckIfFlowMessageIsValid(user, message); err != nil {
		return nil
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
			h.text.EnterDescription,
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
			h.text.EnterPrice,
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
			h.text.EnterCity,
		); err != nil {
			return err
		}

	case models.StateWaitingForCCity:
		user, err := h.db.ChangeAdCity(user, message.Text)

		if err != nil {
			return err
		}

		if !user.Context.Advertisement.Editing {
			if err := h.SendPreview(user); err != nil {
				return err
			}
		}

		if _, err := h.GoNextIfCreatingElseDropEditing(
			user,
			models.StateNONE,
			h.text.AdPreview,
		); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) GoNextIfCreatingElseDropEditing(user *models.User, state models.BotState, messagetext string) (*models.User, error) {
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

	return h.AfterEdited(user)
}

func (h *Handlers) AfterEdited(user *models.User) (*models.User, error) {
	_, err := h.bot.Send(h.CreateNewAdMessage(user, tgbotapi.ModeHTML))
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

func (h *Handlers) CreateNewAdMessage(user *models.User, parsemode string) tgbotapi.MessageConfig {
	username := user.Username

	if DEBUG {
		username = h.text.Hidden
	}

	message := tgbotapi.NewMessage(user.Chatid, formatters.FormatAdToMessageString(user.Context.Advertisement, username))
	message.ParseMode = parsemode
	message.ReplyMarkup = h.GetPreviewMarkup()

	return message
}

func (h *Handlers) HandleStart(user *models.User) error {
	if err := h.SendMessage(user, h.text.Start); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) HandleAddAd(user *models.User) error {
	if err := h.SendMessage(user, h.text.AdGuide); err != nil {
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

func (h *Handlers) SendPreview(user *models.User) error {
	if _, err := h.bot.Send(h.CreateNewAdMessage(user, tgbotapi.ModeHTML)); err != nil {
		return err
	}

	return nil
}

func (h *Handlers) GetPreviewMarkup() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(commands.SendButtonPair.ParamName, (commands.SendButtonPair.ParamValue).(string)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeTitleButton.ParamName, fmt.Sprintf("%s:%d", commands.ChangeTitleButton.ParamValue, models.StateWaitingForCTitle)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeDescriptionButton.ParamName, fmt.Sprintf("%s:%d", commands.ChangeDescriptionButton.ParamValue, models.StateWaitingForCDescription)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangePriceButton.ParamName, fmt.Sprintf("%s:%d", commands.ChangePriceButton.ParamValue, models.StateWaitingForCPrice)),
			tgbotapi.NewInlineKeyboardButtonData(commands.ChangeCityButton.ParamName, fmt.Sprintf("%s:%d", commands.ChangeCityButton.ParamValue, models.StateWaitingForCCity)),
		),
	)
}

func (h *Handlers) HandleCallbackQuery(query *tgbotapi.CallbackQuery) error {
	querydata := strings.Split(query.Data, ":")

	user, err := h.db.GetUser(query.Message.Chat.ID)

	if err != nil {
		return err
	}

	switch querydata[0] {
	case commands.SendButtonPair.ParamValue:
		message := tgbotapi.NewMessageToChannel(h.settings.ManageChannelLink, formatters.FormatAdToMessageString(user.Context.Advertisement, user.Username))
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
		statenum, err := strconv.Atoi(querydata[1])

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

		if err = h.SendMessage(user, h.text.NewParameterValue); err != nil {
			return err
		}
	}

	return nil
}

func (h *Handlers) AskForKey(message *tgbotapi.Message) (*models.User, error) {
	if message.Text == h.settings.SecretKey {
		user, err := h.db.Register(message.Chat.ID, message.From.UserName)
		if err != nil {
			log.Fatal(err)
		}

		message.Text = commands.StartCommand

		return user, nil
	}

	if err := h.SendMessage(models.NewUser(message.Chat.ID, "", nil), h.text.AccessOnlyByKey); err != nil {
		return nil, err
	}

	return nil, nil
}
