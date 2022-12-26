package sqlstorage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ennwy/calendar/internal/app"
	"github.com/ennwy/calendar/internal/storage"
	"github.com/jackc/pgx/v4"
)

const (
	qCreate     = `SELECT NEW_EVENT(owner := $1, title := $2, start := $3, finish := $4, notify := $5);`
	qUpdate     = "UPDATE events SET title = $2, start = $3, finish = $4, notify = $5 WHERE id = $1;"
	qDelete     = "DELETE FROM events WHERE ID = $1;"
	qUserEvents = "SELECT * FROM events WHERE owner = $1;"

	// BETWEEN ISN'T USED BECAUSE WE DON'T NEED NOTIFY PARAM TO BE EQUAL SECOND ARG
	qListUpcoming = "SELECT * FROM events WHERE notify >= $1 AND notify < $2;"

	//qListUsersUpcoming = "SELECT * FROM events WHERE Owner = $1 AND notify >= $2 AND notify < $3"
	qListUsersUpcoming = "SELECT * FROM events WHERE owner = $1 AND notify BETWEEN $2 AND $3;"
	qClean             = "DELETE FROM events WHERE finish <= $1;"

	// Time format for db select
	minuteRounded = "2006-01-02 15:04:05"
)

type Storage struct {
	db     *pgx.Conn
	config *DBConfig
}

var (
	_ app.Storage       = (*Storage)(nil)
	_ app.CleanListener = (*Storage)(nil)
)

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
	if s.db, err = pgx.Connect(ctx, s.config.getConnectString()); err != nil {
		return fmt.Errorf("storage connect: %w", err)
	}
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if err := s.db.Close(ctx); err != nil {
		return fmt.Errorf("storage close: %w", err)
	}

	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, e *storage.Event) error {
	row := s.db.QueryRow(
		ctx,
		qCreate,

		e.Owner.Name,
		e.Title,
		e.Start,
		e.Finish,
		e.GetNotifyTime(),
	)
	err := row.Scan(&e.ID)
	l.Info("Created EVENT: ID:", e.ID, "; Owner:", e.Owner.Name)

	return err
}

func (s *Storage) UpdateEvent(ctx context.Context, e *storage.Event) error {
	_, err := s.db.Exec(
		ctx,
		qUpdate,

		e.ID,
		e.Title,
		e.Start,
		e.Finish,
		e.GetNotifyTime(),
	)

	return err
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID int64) error {
	if _, err := s.db.Exec(ctx, qDelete, eventID); err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

func (s *Storage) ListUserEvents(ctx context.Context, user string) (*storage.Events, error) {
	rows, err := s.db.Query(ctx, qUserEvents, user)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	return getEvents(rows), nil
}

func (s *Storage) ListUpcoming(ctx context.Context, until time.Duration) (*storage.Events, error) {
	if until < 0 {
		until = -until
	}
	current := time.Now().Round(time.Minute)
	bound := current.Add(until)

	l.Info("Looking for events between:", current.Format(minuteRounded), "AND", bound.Format(minuteRounded))

	rows, err := s.db.Query(
		ctx,
		qListUpcoming,
		current.Format(minuteRounded),
		bound.Format(minuteRounded),
	)

	if err != nil {
		return nil, fmt.Errorf("list upcoming: %w", err)
	}

	return getEvents(rows), err
}

func (s *Storage) ListUsersUpcoming(ctx context.Context, user string, until time.Duration) (*storage.Events, error) {
	if until < 0 {
		until = -until
	}

	current := time.Now()
	bound := current.Add(until)

	l.Info("Looking for events between:", current.Format(minuteRounded), "AND", bound.Format(minuteRounded))

	rows, err := s.db.Query(
		ctx,
		qListUsersUpcoming,
		user,
		current.Format(minuteRounded),
		bound.Format(minuteRounded),
	)

	if err != nil {
		return nil, fmt.Errorf("list upcoming: %w", err)
	}

	return getEvents(rows), err
}

func (s *Storage) Clean(ctx context.Context, ago time.Duration) error {
	if ago > 0 {
		ago = -ago
	}

	finishedAgo := time.Now().UTC().Round(storage.Day).Add(ago)
	l.Info("[ + ] Scheduler CLEAN ago:", finishedAgo.Format(minuteRounded))

	if _, err := s.db.Exec(
		ctx,
		qClean,
		finishedAgo.Format(minuteRounded),
	); err != nil {
		return fmt.Errorf("clean events: %w", err)
	}

	return nil
}

func getEvents(rows pgx.Rows) *storage.Events {
	events := storage.Events{
		Events: make([]storage.Event, 0, 1),
	}

	var notifyTime time.Time
	var e storage.Event

	for rows.Next() {
		if err := rows.Scan(
			&e.ID,
			&e.Owner.Name,
			&e.Owner.ID,
			&e.Title,
			&e.Start,
			&e.Finish,
			&notifyTime,
		); err != nil {
			l.Error("sql storage listing: ", err)
			continue
		}

		e.SetNotifyByTime(notifyTime)
		l.Info("Notify from base:", e.Notify)
		events.Events = append(events.Events, e)
		e.Reset()
	}

	return &events
}
