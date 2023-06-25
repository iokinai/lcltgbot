package lcltgbot

import (
	"database/sql"
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

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS temp_ads (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, title VARCHAR(255), description TEXT, price DOUBLE, city TEXT)"); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS temp_contexts (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, is_in_flow INTEGER, ad_id INTEGER, state INTEGER, FOREIGN KEY(ad_id) REFERENCES temp_ads(id))"); err != nil {
		log.Fatal(err)
	}

	if _, err := db.Exec("CREATE TABLE users (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, chat_id INTEGER UNIQUE, context_id INTEGER, FOREIGN KEY(context_id) REFERENCES temp_contexts(id))"); err != nil {
		log.Fatal(err)
	}

	return &SqliteDb{db: db}
}

func (s *SqliteDb) Register(chatid int64) (*models.User, error) {
	_, err := s.db.Exec("INSERT INTO users(chat_id, context_id) VALUES (?, ?)", chatid, nil)

	if err != nil {
		return nil, err
	}

	return models.NewUser(chatid, nil), nil
}

func (s *SqliteDb) GetUser(chatid int64) (*models.User, error) {
	userrows, err := s.GetRowsById("SELECT * FROM users WHERE chat_id = ?", chatid)

	if err != nil {
		return nil, err
	}

	for userrows.Next() {
		//TODO: доделать
	}
}

func (s *SqliteDb) GetRowsById(query string, id int64) (*sql.Rows, error) {
	rows, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}

	return rows, nil
}
