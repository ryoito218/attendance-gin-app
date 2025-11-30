package domain

import "time"

func CalcWorkDuration(clockIn, clockOut time.Time) time.Duration {
	if clockOut.Before(clockIn) {
		return 0
	}
	return clockOut.Sub(clockIn)
}