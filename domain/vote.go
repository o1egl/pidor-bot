package domain

import "time"

type Vote struct {
	UserID      int64
	VotedUserID int64
	Time        time.Time
}
