package provider

import (
	"fmt"
	"github.com/saddlemc/launcher/plugin"
	"os"
	"path/filepath"
)

type LocalPlugin struct {
	identifier plugin.Identifier
	path       string
}

func LocalProvider(info map[string]any) (plugin.Plugin, error) {
	p, ok := info["local"]
	if !ok {
		return nil, nil
	}
	path, ok := p.(string)
	if !ok {
		return nil, nil
	}

	path, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		panic(err)
	}
	identifier, err := plugin.ParseIdentifier(path)
	if err != nil {
		return nil, err
	}
	return &LocalPlugin{
		identifier: identifier,
		path:       path,
	}, nil
}

func (l *LocalPlugin) Latest() (plugin.Identifier, error) {
	return l.identifier, nil
}

func (l *LocalPlugin) Pull() error {
	if _, err := os.Stat(l.path); os.IsNotExist(err) {
		return fmt.Errorf("path '%w' does not exist", l.path)
	}
	// Nothing else needs to be done as the plugin is already locally downloaded.
	return nil
}

func (l *LocalPlugin) Module() plugin.Module {
	return plugin.Module{
		Module:  l.identifier.Module,
		Version: "v0.0.0",
		Replace: l.path,
		Import:  l.identifier.Module + "/import",
	}
}
