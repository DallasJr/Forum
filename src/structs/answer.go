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
	PostTitle    string
	CreationDate string
	Creator      User
	Likes        []uuid.UUID `json:"likes"`
	Dislikes     []uuid.UUID `json:"dislikes"`
}

func (a *Answer) FormattedDate() (string, error) {
	inputLayout := "2006-01-02 15:04:05"

	t, err := time.Parse(inputLayout, a.CreationDate)
	if err != nil {
		return "", err
	}

	dateOutputLayout := "January 2, 2006"
	timeOutputLayout := "3:04pm"

	dateStr := t.Format(dateOutputLayout)
	timeStr := t.Format(timeOutputLayout)

	formattedDateTime := dateStr + " " + timeStr

	return formattedDateTime, nil
}
