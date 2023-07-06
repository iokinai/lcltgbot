package commands

import "github.com/iokinai/lcltgbot/internal/lcltgbot/models"

const (
	StartCommand = "/start"
	CancelFlow   = "/cancel_flow"
	AddAdCommand = "/add_ad"
)

const ChangeValueCommandData = "changevalue"

var (
	SendButtonPair          = models.NewParamPair("Отправить", "send")
	ChangeTitleButton       = models.NewParamPair("Изменить заголовок", ChangeValueCommandData)
	ChangeDescriptionButton = models.NewParamPair("Изменить описание", ChangeValueCommandData)
	ChangePriceButton       = models.NewParamPair("Изменить цену", ChangeValueCommandData)
	ChangeCityButton        = models.NewParamPair("Изменить город", ChangeValueCommandData)
)
