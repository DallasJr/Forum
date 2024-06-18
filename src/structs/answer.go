package structs

import (
	"github.com/google/uuid"
	"time"
)

type Answer struct {
	Uuid         uuid.UUID
	Content      string
	CreatorUUID  uuid.UUID
	PostID       uuid.UUID
	CreationDate string
	Creator      User
	Likes        []uuid.UUID `json:"likes"`
	Dislikes     []uuid.UUID `json:"dislikes"`
}

func (a *Answer) FormattedDate() (string, error) {
	// Define the layout of the input date string
	inputLayout := "2006-01-02 15:04:05"

	// Parse the date string into a time.Time object
	t, err := time.Parse(inputLayout, a.CreationDate)
	if err != nil {
		return "", err
	}

	// Define the desired output layouts
	dateOutputLayout := "January 2, 2006"
	timeOutputLayout := "3:04pm"

	// Format the time.Time object into the desired string formats
	dateStr := t.Format(dateOutputLayout)
	timeStr := t.Format(timeOutputLayout)

	// Combine date and time strings
	formattedDateTime := dateStr + " " + timeStr

	return formattedDateTime, nil
}
