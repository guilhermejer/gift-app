package domain

import "time"

type User struct {
	UserID    string
	FullName  string
	Email     string
	Active    bool
	PlanID    string
	BirthDate time.Time
	City      string
}
