package backend

import "time"

const lockTTL = 30 * time.Minute

type lock struct {
	value  string
	expiry time.Time
}
