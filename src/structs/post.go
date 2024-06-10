package structs

import (
	"github.com/google/uuid"
	"time"
)

type Post struct {
	Uuid         uuid.UUID
	Title        string
	Content      string
	Creator      uuid.UUID
	Category     string
	CreationDate time.Time
	Likes        []uuid.UUID
	Dislikes     []uuid.UUID
}
