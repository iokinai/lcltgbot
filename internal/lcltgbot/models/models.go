package models

type BotData struct {
	Key string
}

func NewBotData(key string) *BotData {
	return &BotData{Key: key}
}

type Advertisement struct {
	Title       string
	Description string
	Price       float64
	City        string
}

func NewAdvertisement(title string, description string, price float64, city string) *Advertisement {
	return &Advertisement{Title: title, Description: description, Price: price, City: city}
}

type BotState int8

const (
	StateNONE BotState = iota
	StateWaitingForCTitle
	StateWaitingForCDescription
	StateWaitingForCPrice
	StateWaitingForCCity
)

type BotContext struct {
	IsInFlow      bool
	Advertisement *Advertisement
	State         BotState
}

type User struct {
	Chatid  int64
	Context *BotContext
}

func NewUser(chatid int64, context *BotContext) *User {
	return &User{Chatid: chatid, Context: context}
}
