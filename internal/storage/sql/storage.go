package sqlstorage

import (
	"context"
	"fmt"
	"os"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/jackc/pgx/v4"
)

const (
	dbSelect = "SELECT * FROM events WHERE ownerID = $1;"
	dbInsert = "INSERT INTO events(startTime, finishTime, title, ownerId) VALUES ($1, $2, $3, $4) RETURNING id;"
	dbUpdate = "UPDATE events SET startTime = $2, finishTime = $3, title = $4 WHERE ID = $1;"
	dbDelete = "DELETE FROM events WHERE ID = $1;"
)

type Storage struct {
	db     *pgx.Conn
	config *DBConfig
}

var _ app.Storage = (*Storage)(nil)

var l app.Logger

func New(logger app.Logger) *Storage {
	l = logger

	s := &Storage{
		config: NewDBConf(),
	}

	return s
}

type DBConfig struct {
	port     string
	host     string
	name     string
	user     string
	password string
}

func (db *DBConfig) getConnectString() string {
	info := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.host,
		db.port,
		db.user,
		db.password,
		db.name,
	)

	l.Info("[ + ] Database Config: ", info)

	return info
}

func NewDBConf() *DBConfig {
	return &DBConfig{
		host:     os.Getenv("DATABASE_HOST"),
		port:     os.Getenv("DATABASE_PORT"),
		user:     os.Getenv("DATABASE_USER"),
		password: os.Getenv("DATABASE_PASSWORD"),
		name:     os.Getenv("DATABASE_NAME"),
	}
}

func (s *Storage) Connect(ctx context.Context) (err error) {
	s.db, err = pgx.Connect(ctx, s.config.getConnectString())
	return err
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close(ctx)
}

func (s *Storage) CreateEvent(ctx context.Context, e *storage.Event) error {
	row := s.db.QueryRow(
		ctx,
		dbInsert,

		e.Start,
		e.Finish,
		e.Title,
		e.OwnerID,
	)

	return row.Scan(&e.ID)
}

func (s *Storage) ListEvents(ctx context.Context, ownerID int64) ([]storage.Event, error) {
	rows, err := s.db.Query(ctx, dbSelect, ownerID)
	if err != nil {
		return nil, err
	}

	eventList := make([]storage.Event, 0, 1)

	var e storage.Event

	for rows.Next() {
		err = rows.Scan(&e.ID, &e.OwnerID, &e.Start, &e.Finish, &e.Title)
		if err != nil {
			l.Error("sql storage  listing: ", err)
			continue
		}

		eventList = append(eventList, e)
	}

	return eventList, err
}

func (s *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	_, err := s.db.Exec(
		ctx,
		dbUpdate,

		e.ID,
		e.Start,
		e.Finish,
		e.Title,
	)

	return err
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID int64) error {
	_, err := s.db.Exec(ctx, dbDelete, eventID)
	return err
}
