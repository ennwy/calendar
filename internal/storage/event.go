package storage

import (
	"github.com/ghodss/yaml"
	"time"
)

const (
	Day = time.Hour * 24
)

type User struct {
	Name string
	ID   int64
}

type Event struct {
	Start  time.Time `yaml:"start"`
	Finish time.Time `yaml:"finish"`
	Owner  User      `yaml:"owner"`
	Title  string    `yaml:"title"`
	ID     int64     `yaml:"id"`
	/*
		Notify: Number of minutes you need to notify before event Start.
		Value always negative.
	*/
	Notify int32 `yaml:"notify"`
}

func (e *Event) GetNotifyTime() time.Time {
	m := time.Minute
	start := e.Start.Round(m)
	return start.Add(time.Duration(-e.Notify) * m)
}

func (e *Event) SetNotifyByTime(t time.Time) {
	e.Notify = int32(e.Start.Sub(t) / 60 / 1000 / 1000 / 1000)
}

func (e *Event) Reset() {
	*e = Event{}
}

func (e *Event) Marshall() ([]byte, error) {
	return yaml.Marshal(e)
}

func (e *Event) Unmarshall(EventYaml []byte) error {
	return yaml.Unmarshal(EventYaml, e)
}
