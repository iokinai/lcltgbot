package lcltgbot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/iokinai/lcltgbot/internal/lcltgbot/models"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type SqliteDb struct {
	db *sql.DB
}

func NewSqliteDb() *SqliteDb {
	db, err := sql.Open("sqlite3", "lcl.sqlite")

	if err != nil {
		log.Fatal(err)
	}

	CreateTableOrError(db, "temp_ads (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, title VARCHAR(255), description TEXT, price DOUBLE, city TEXT, editing BOOLEAN)")
	CreateTableOrError(db, "temp_contexts (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, is_in_flow INTEGER, ad_id INTEGER, state INTEGER, FOREIGN KEY(ad_id) REFERENCES temp_ads(id))")
	CreateTableOrError(db, "users (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, chat_id INTEGER UNIQUE, context_id INTEGER, FOREIGN KEY(context_id) REFERENCES temp_contexts(id))")

	return &SqliteDb{db: db}
}

func CreateTableOrError(db *sql.DB, data string) {
	if _, err := db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s", data)); err != nil {
		log.Fatal(err)
	}
}

func (s *SqliteDb) Register(chatid int64) (*models.User, error) {
	emptyad, err := s.CreateAd("", "", 0, "")

	if err != nil {
		return nil, err
	}

	emptyctx, err := s.CreateContext(false, emptyad, models.StateNONE)

	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec("INSERT INTO users(chat_id, context_id) VALUES (?, ?)", chatid, emptyctx.Id)

	if err != nil {
		return nil, err
	}

	return models.NewUser(chatid, emptyctx), nil
}

func (s *SqliteDb) CreateContext(isInFlow bool, ad *models.Advertisement, state models.BotState) (*models.BotContext, error) {
	result, err := s.db.Exec("INSERT INTO temp_contexts(is_in_flow, ad_id, state) VALUES (?, ?, ?)", isInFlow, ad.Id, state)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &models.BotContext{Id: id, IsInFlow: isInFlow, Advertisement: ad, State: state}, nil
}

func (s *SqliteDb) CreateAd(title string, description string, price float64, city string) (*models.Advertisement, error) {
	result, err := s.db.Exec("INSERT INTO temp_ads(title, description, price, city, editing) VALUES (?, ?, ?, ?, ?)", title, description, price, city, false)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return models.NewAdvertisement(id, title, description, price, city, false), nil
}

func (s *SqliteDb) GetUser(chatid int64) (*models.User, error) {
	userrows, err := s.GetRowsById("SELECT * FROM users WHERE chat_id = ?", chatid)

	loaded := false

	if err != nil {
		return nil, err
	}

	var (
		userId     int
		chatId     int
		ucontextId sql.NullInt64
	)

	for userrows.Next() {
		if err := userrows.Scan(&userId, &chatId, &ucontextId); err != nil {
			return nil, err
		}
		loaded = true
	}

	if !loaded {
		return nil, errors.New("no values in DB")
	}

	context, err := s.GetContext(ucontextId)

	if err != nil {
		return nil, err
	}

	return models.NewUser(chatid, context), nil
}

func (s *SqliteDb) GetAd(id sql.NullInt64) (*models.Advertisement, error) {
	if !id.Valid {
		return nil, nil
	}

	loaded := false

	var (
		adId        int64
		title       string
		description string
		price       float64
		city        string
		editing     bool
	)

	adrows, err := s.GetRowsById("SELECT * FROM temp_ads WHERE id = ?", id.Int64)

	if err != nil {
		return nil, err
	}

	for adrows.Next() {
		if err := adrows.Scan(&adId, &title, &description, &price, &city, &editing); err != nil {
			return nil, err
		}
		loaded = true
	}

	if !loaded {
		return nil, errors.New("no values in DB")
	}

	return models.NewAdvertisement(adId, title, description, price, city, editing), nil
}

func (s *SqliteDb) GetContext(id sql.NullInt64) (*models.BotContext, error) {
	if !id.Valid {
		return nil, nil
	}

	loaded := false

	var (
		contextId int64
		isInFlow  bool
		uadId     sql.NullInt64
		state     int
	)

	contextrows, err := s.GetRowsById("SELECT * FROM temp_contexts WHERE id = ?", id.Int64)

	if err != nil {
		return nil, err
	}

	for contextrows.Next() {
		if err := contextrows.Scan(&contextId, &isInFlow, &uadId, &state); err != nil {
			return nil, err
		}
		loaded = true
	}

	if !loaded {
		return nil, errors.New("no values in DB")
	}

	ad, err := s.GetAd(uadId)

	if err != nil {
		return nil, err
	}

	return &models.BotContext{
		Id:            contextId,
		IsInFlow:      isInFlow,
		Advertisement: ad,
		State:         models.BotState(state),
	}, nil
}

func (s *SqliteDb) GetRowsById(query string, id int64) (*sql.Rows, error) {
	rows, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *SqliteDb) ChangeUserState(user *models.User, state models.BotState) (*models.User, error) {
	user.Context.IsInFlow = true

	if state == models.StateNONE {
		user.Context.IsInFlow = false
	}

	user.Context.State = state

	_, err := s.db.Exec("UPDATE temp_contexts SET is_in_flow = ?, state = ? WHERE id = ?", user.Context.IsInFlow, user.Context.State, user.Context.Id)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *SqliteDb) ChangeAdTitle(user *models.User, title string) (*models.User, error) {
	user.Context.Advertisement.Title = title
	return user, s.ChangeAdParam(user, title, "title")
}

func (s *SqliteDb) ChangeAdDescription(user *models.User, descr string) (*models.User, error) {
	user.Context.Advertisement.Description = descr
	return user, s.ChangeAdParam(user, descr, "description")
}

func (s *SqliteDb) ChangeAdPrice(user *models.User, price float64) (*models.User, error) {
	user.Context.Advertisement.Price = price
	return user, s.ChangeAdParam(user, price, "price")
}

func (s *SqliteDb) ChangeAdCity(user *models.User, city string) (*models.User, error) {
	user.Context.Advertisement.City = city
	return user, s.ChangeAdParam(user, city, "city")
}

func (s *SqliteDb) ChangeAdEditing(user *models.User, editing bool) (*models.User, error) {
	user.Context.Advertisement.Editing = editing
	return user, s.ChangeAdParam(user, editing, "editing")
}

func (s *SqliteDb) ChangeAdParam(user *models.User, param any, paramname string) error {
	_, err := s.db.Exec(fmt.Sprintf("UPDATE temp_ads SET %v = ? WHERE id = ?", paramname), param, user.Context.Advertisement.Id)

	if err != nil {
		return err
	}

	return nil
}
