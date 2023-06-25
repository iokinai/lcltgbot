package handlers

import "github.com/iokinai/lcltgbot/internal/lcltgbot/models"

type Handlers struct {
	currentUser *models.User
}

func (h *Handlers) HandleSingleCommand(user *models.User) error {

}

func (h *Handlers) HandleCommandFlow(user *models.User) error {

}
