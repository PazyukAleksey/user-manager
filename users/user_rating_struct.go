package users

import "time"

type UserRatingList struct {
	VotedNickname string
	VotedRating   int
	VotedDate     time.Time
}
