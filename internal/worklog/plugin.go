package worklog

import (
	"daiv/internal/plugin"
	"fmt"
)

type WorklogPlugin struct {
	WorklogPath string
}

func NewWorklogPlugin() *WorklogPlugin {
	return &WorklogPlugin{}
}

func (g *WorklogPlugin) Name() string {
	return "worklog"
}

func (g *WorklogPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{
				Type:        plugin.ConfigTypeString,
				Key:         "worklog.path",
				Name:        "Worklog Path",
				Description: "The path to the worklog file",
				Required:    true,
			},
		},
	}
}

func (g *WorklogPlugin) Initialize(settings map[string]any) error {
	worklogPath := settings["worklog.path"].(string)
	if worklogPath == "" {
		return fmt.Errorf("worklog.path is required")
	}

	g.WorklogPath = worklogPath

	return nil
}

func (g *WorklogPlugin) Shutdown() error {
	return nil
}

func (g *WorklogPlugin) GetStandupContext(timeRange plugin.TimeRange) (plugin.StandupContext, error) {
	standupContext := NewStandupContext(g.WorklogPath, timeRange)
	content, err := standupContext.Render()
	if err != nil {
		return plugin.StandupContext{}, err
	}

	return plugin.StandupContext{
		PluginName: g.Name(),
		Content:    content,
	}, nil
}
