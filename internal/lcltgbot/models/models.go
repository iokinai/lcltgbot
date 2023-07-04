package models

type BotData struct {
	Key       string `json:"key"`
	SecretKey string `json:"secretKey"`
}

func NewBotData(key string) *BotData {
	return &BotData{Key: key}
}

type Advertisement struct {
	Id          int64
	Title       string
	Description string
	Price       float64
	City        string
}

func NewAdvertisement(id int64, title string, description string, price float64, city string) *Advertisement {
	return &Advertisement{Id: id, Title: title, Description: description, Price: price, City: city}
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
	Id            int64
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
