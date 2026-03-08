package api

import "time"

func now() *time.Time {
	t := time.Now().UTC()
	return &t
}
