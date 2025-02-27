package utils

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

func IsDateTimeInThreshold(timestamp time.Time) bool {
	fromTime, err := time.Parse(time.RFC3339, viper.GetString("fromTime"))
	if err != nil {
		return false
	}

	toTime, err := time.Parse(time.RFC3339, viper.GetString("toTime"))
	if err != nil {
		return false
	}

	return timestamp.After(fromTime) && timestamp.Before(toTime)
}

func IsDateOnOrAfter(timestamp time.Time, threshold time.Time) bool {
	return timestamp.After(threshold) || timestamp.Equal(threshold)
}

func GetCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	daivDir := filepath.Join(cacheDir, "daiv")
	if err := os.MkdirAll(daivDir, 0755); err != nil {
		return "", err
	}
	return daivDir, nil
}
