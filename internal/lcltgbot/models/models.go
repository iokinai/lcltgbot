package models

type AppSettings struct {
	Key               string `json:"key"`
	SecretKey         string `json:"secretKey"`
	ManageChannelLink string `json:"manageChannelLink"`
	DatabasePath      string `json:"databasePath"`
}

func NewBotData(key string) *AppSettings {
	return &AppSettings{Key: key}
}

type Advertisement struct {
	Id          int64
	Title       string
	Description string
	Price       float64
	City        string
	Editing     bool
}

func NewAdvertisement(id int64, title string, description string, price float64, city string, editing bool) *Advertisement {
	return &Advertisement{Id: id, Title: title, Description: description, Price: price, City: city, Editing: editing}
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
	Chatid   int64
	Username string
	Context  *BotContext
}

func NewUser(chatid int64, username string, context *BotContext) *User {
	return &User{Chatid: chatid, Username: username, Context: context}
}

type ParamPair struct {
	ParamName  string
	ParamValue any
}

func NewParamPair(paramName string, paramValue any) *ParamPair {
	return &ParamPair{ParamName: paramName, ParamValue: paramValue}
}

type TextSettings struct {
	Start             string `json:"start"`
	WrongCommand      string `json:"wrongCommand"`
	InChainError      string `json:"inChainError"`
	ChainCanceled     string `json:"chainCanceled"`
	AdGuide           string `json:"adGuide"`
	EnterDescription  string `json:"enterDescription"`
	EnterPrice        string `json:"enterPrice"`
	EnterCity         string `json:"enterCity"`
	AdPreview         string `json:"adPreview"`
	Hidden            string `json:"hidden"`
	NewParameterValue string `json:"newParameterValue"`
	AccessOnlyByKey   string `json:"accessOnlyByKey"`
}
