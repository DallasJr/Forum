package structs

import "time"

type User struct {
	Surname      string
	Name         string
	Email        string
	Password     string
	CreationDate time.Time
	Admin        bool
	Gender       bool
}
