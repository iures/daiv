package worklog

import (
	"daiv/pkg/plugin"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StandupContext struct {
	worklogPath string
	Content string
	timeRange plugin.TimeRange
}

func NewStandupContext(worklogPath string, timeRange plugin.TimeRange) *StandupContext {
	return &StandupContext{
		worklogPath: worklogPath,
		timeRange:   timeRange,
	}
}

func (s *StandupContext) Render() (string, error) {
	if s.worklogPath == "" {
		return "", nil
	}

	items, err := s.fetchWorkItems()
	if err != nil {
		return "", fmt.Errorf("error loading worklog: %v", err)
	}

	var report strings.Builder

	for _, item := range items {
		if s.timeRange.IsInRange(item.Timestamp) {
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

type WorkItem struct {
	Contents    string
	Timestamp   time.Time
	Tags        []string
}


func (s *StandupContext) fetchWorkItems() ([]WorkItem, error) {
	var items []WorkItem

	dir := filepath.Dir(s.worklogPath)

	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, fmt.Errorf("error getting files: %v", err)
	}

	for _, filePath := range files {
		info, err := os.Stat(filePath)
		if err != nil {
			continue // Skip files which we cannot stat.
		}

		if s.timeRange.IsInRange(info.ModTime()) {
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
