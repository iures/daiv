package standup

import (
	"bakuri/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type WorkItem struct {
	Contents    string
	Timestamp   time.Time
	Tags        []string
}

type WorklogReport struct {
	worklogPath string
}

func NewWorklogReport() *WorklogReport {
	return &WorklogReport{
		worklogPath: viper.GetString("worklog.path"),
	}
}

func (w *WorklogReport) Render() (string, error) {
	if w.worklogPath == "" {
		return "", nil
	}

	items, err := w.loadWorkItems()
	if err != nil {
		return "", fmt.Errorf("error loading worklog: %v", err)
	}

	var report strings.Builder

	for _, item := range items {
		if utils.IsDateTimeInThreshold(item.Timestamp) {
			tagStr := ""
			if len(item.Tags) > 0 {
				tagStr = fmt.Sprintf(" [%s]", strings.Join(item.Tags, ", "))
			}
			report.WriteString(fmt.Sprintf("- %s%s (Logged: %s)\n", 
				item.Contents,
				tagStr,
				item.Timestamp.Format("2006-01-02 15:04:05")))
		}
	}

	return report.String(), nil
}

func (w *WorklogReport) loadWorkItems() ([]WorkItem, error) {
	var items []WorkItem

	dir := filepath.Dir(w.worklogPath)

	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, fmt.Errorf("error getting files: %v", err)
	}

	threshold := time.Now().AddDate(0, 0, -1)

	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			continue // Skip files which we cannot stat.
		}

		if info.ModTime().After(threshold) || info.ModTime().Equal(threshold) {
			contents, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			items = append(items, WorkItem{
				Contents:    string(contents),
				Timestamp:   info.ModTime(),
			})
		}
	}

	return items, nil
}
