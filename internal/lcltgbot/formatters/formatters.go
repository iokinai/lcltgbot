package formatters

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
)

const DEBUG = true

func FormatAdToMessageString(advertisement *models.Advertisement, username string) string {
	debugmessage := ""

	if DEBUG {
		debugmessage = `
<b>*юзернеймы пользователей скрыты в бета версии</b>
`
	}

	return fmt.Sprintf(
		`<b>%s</b>


<i>Описание:</i>
%s

Цена: <b>%s ₽</b>

<i>г. %s</i>

Писать в: %s
%s			
	`, advertisement.Title, advertisement.Description, humanize.FormatFloat("# ###.##", advertisement.Price), advertisement.City, username, debugmessage)
}
