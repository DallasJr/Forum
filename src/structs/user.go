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
	// Define the layout of the input date string
	inputLayout := "2006-01-02 15:04:05"

	// Parse the date string into a time.Time object
	t, err := time.Parse(inputLayout, u.CreationDate)
	if err != nil {
		return "", err
	}

	// Define the desired output layouts
	dateOutputLayout := "January 2, 2006"

	// Format the time.Time object into the desired string formats
	formattedDate := t.Format(dateOutputLayout)

	return formattedDate, nil
}
