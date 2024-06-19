package structs

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Uuid         uuid.UUID
	Title        string
	Content      string
	CreatorUUID  uuid.UUID
	Category     string
	CreationDate string
	Creator      User
	Likes        []uuid.UUID `json:"likes"`
	Dislikes     []uuid.UUID `json:"dislikes"`
	Images       []string    `json:"images"`
	AnswersCount int
}

func (p *Post) FormattedDate() (string, error) {
	inputLayout := "2006-01-02 15:04:05"

	t, err := time.Parse(inputLayout, p.CreationDate)
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

func Shorten(word string, size int) string {
	if len(word) > size {
		word = word[:size] + "..."
	}
	return word
}
