package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id"`
	FirstName      string
	LastName       string
	Nickname       string
	Email          string
	Password       string
	Information    string
	Role           string
	Rating         int
	UserRatingList string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
	VotedAt        time.Time
}
