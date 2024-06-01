package structs

import (
	"github.com/google/uuid"
	"time"
)

type Answer struct {
	Uuid         uuid.UUID
	Content      string
	Creator      uuid.UUID
	PostID       uuid.UUID
	CreationDate time.Time
	Likes        []uuid.UUID
	Dislikes     []uuid.UUID
}
