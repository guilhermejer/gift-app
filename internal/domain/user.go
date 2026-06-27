package domain

import "time"

type User struct {
	UserID    string
	Active    bool
	PlanID    string
	BirthDate time.Time
	City      string
}
