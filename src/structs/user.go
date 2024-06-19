package structs

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Uuid         uuid.UUID
	Name         string
	Surname      string
	Username     string
	Email        string
	Password     string
	Gender       string
	CreationDate string
	Power        int
}

func (u *User) FormattedDate() (string, error) {
	inputLayout := "2006-01-02 15:04:05"

	t, err := time.Parse(inputLayout, u.CreationDate)
	if err != nil {
		return "", err
	}

	dateOutputLayout := "January 2, 2006"

	formattedDate := t.Format(dateOutputLayout)

	return formattedDate, nil
}
