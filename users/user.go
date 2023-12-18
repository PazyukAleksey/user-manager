package users

import (
	"time"
)

type User struct {
	FirstName   string
	LastName    string
	Nickname    string
	Email       string
	Password    string
	Information string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}
