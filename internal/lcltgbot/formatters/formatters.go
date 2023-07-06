package formatters

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
)

func FormatAdToMessageString(advertisement *models.Advertisement, username string) string {
	return fmt.Sprintf(
		`<b>%s</b>


<i>Описание:</i>
%s

Цена: <b>%s ₽</b>

<i>г. %s</i>

Писать в: %s
			
	`, advertisement.Title, advertisement.Description, humanize.FormatFloat("# ###.##", advertisement.Price), advertisement.City, username)
}
