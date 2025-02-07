package utils

import "time"

func IsDateOnOrAfter(timestamp time.Time, threshold time.Time) bool {
	return timestamp.After(threshold) || timestamp.Equal(threshold)
}
