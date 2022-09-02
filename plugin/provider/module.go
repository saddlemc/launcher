package provider

import (
	"errors"
	"github.com/saddlemc/launcher/plugin"
)

type ModulePlugin struct {
	name, version string
}

func ModuleProvider(info map[string]any) (plugin.Plugin, error) {
	n, ok := info["module"]
	if !ok {
		return nil, nil
	}
	name, ok := n.(string)
	if !ok {
		return nil, nil
	}
	version := "latest"
	if x, ok := info["version"]; ok {
		v, ok := x.(string)
		if !ok {
			return nil, errors.New("plugin version must be surrounded by \"\"")
		}
		version = v
	}
	// todo: convert tag into commit hash instead of abusing go modules
	return ModulePlugin{
		name:    name,
		version: version,
	}, nil
}

func (m ModulePlugin) Latest() (plugin.Identifier, error) {
	return plugin.Identifier{
		Module:   m.name,
		Checksum: "git:" + m.version,
	}, nil
}

func (m ModulePlugin) Pull() error {
	// The module will be downloaded by go modules.
	return nil
}

func (m ModulePlugin) Module() plugin.Module {
	return plugin.Module{
		Module:  m.name,
		Version: m.version,
		Replace: "",
		Import:  m.name + "/import",
	}
}
