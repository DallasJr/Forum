package structs

import "github.com/google/uuid"

type Categorie struct {
	Name        string
	Description string
	Posts       []uuid.UUID
}
