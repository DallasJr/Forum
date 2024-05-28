package structs

import "github.com/google/uuid"

type Categories struct {
	Name        string
	Description string
	Posts       []uuid.UUID
}
