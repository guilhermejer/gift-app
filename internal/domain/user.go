package domain

import "time"

const (
	Lookahead7Days  = 7
	Lookahead14Days = 14
	Lookahead30Days = 30
	DefaultLookaheadDays = Lookahead7Days
)

func IsValidLookaheadDays(d int) bool {
	return d == Lookahead7Days || d == Lookahead14Days || d == Lookahead30Days
}

type User struct {
	UserID                 string    `json:"userID" example:"5581a365-394f-467d-ae13-0d01e4cf1863"`
	FullName               string    `json:"fullName" example:"Guilherme Jeronymo"`
	Email                  string    `json:"email" example:"guilherme.jer1@gmail.com"`
	Active                 bool      `json:"active" example:"true"`
	PlanID                 string    `json:"planID" example:""`
	BirthDate              time.Time `json:"birthDate" format:"date-time" example:"1999-03-16T00:00:00Z"`
	City                   string    `json:"city" example:"Santo Andre"`
	SuggestionLookaheadDays int      `json:"suggestionLookaheadDays" example:"7"`
}
