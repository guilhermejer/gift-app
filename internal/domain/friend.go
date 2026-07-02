package domain

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Friend struct {
	FriendID     string    `json:"friendID" example:"9b02ce54-4f42-4a8b-a539-5b53a6e37e63"`
	UserID       string    `json:"userID" example:"5581a365-394f-467d-ae13-0d01e4cf1863"`
	UserRelation string    `json:"userRelation" example:"irma"`
	Name         string    `json:"name" example:"Mariana Souza"`
	Avatar       string    `json:"avatar" example:"🧑‍💻"`
	Gender       Gender    `json:"gender" example:"female"`
	BirthDate    time.Time `json:"birthDate" format:"date-time" example:"1994-10-03T00:00:00Z"`
	City         string    `json:"city" example:"Belo Horizonte"`
	Profile      *Profile  `json:"profile"`
}
