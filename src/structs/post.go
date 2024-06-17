package structs

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Uuid         uuid.UUID
	Title        string
	Content      string
	Creator      string
	Category     string
	CreationDate string
	Likes        []uuid.UUID `json:"likes"` // Assuming likes are stored as a JSON array of UUIDs
	Dislikes     []uuid.UUID `json:"dislikes"`
	Images       []string    `json:"images"`
}

func (p *Post) FormattedDate() (string, error) {
	// Define the layout of the input date string
	inputLayout := "2006-01-02 15:04:05"

	// Parse the date string into a time.Time object
	t, err := time.Parse(inputLayout, p.CreationDate)
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

func Shorten(word string, size int) string {
	if len(word) > size {
		word = word[:size] + "..."
	}
	return word
}
