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
	Gender       bool
	CreationDate time.Time
	Admin        bool
}
