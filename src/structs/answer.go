package structs

import (
	"github.com/google/uuid"
	"time"
)

type Answer struct {
	Uuid         uuid.UUID
	Creator      uuid.UUID
	Likes        []uuid.UUID
	Dislikes     []uuid.UUID
	CreationDate time.Time
}
