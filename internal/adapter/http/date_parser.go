package http

import "time"

const dateOnlyLayout = "2006-01-02"

func parseDateOnly(value string) (time.Time, error) {
	return time.Parse(dateOnlyLayout, value)
}
