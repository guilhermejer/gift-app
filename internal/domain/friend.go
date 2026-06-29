package domain

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Friend struct {
	FriendID     string
	UserID       string
	UserRelation string
	Name         string
	Gender       Gender
	BirthDate    time.Time
	City         string
	Profile      *Profile
}
