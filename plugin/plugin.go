package plugin

import (
	"github.com/kr/pretty"
	"github.com/saddlemc/launcher/config"
)

// Plugin represents a source location for a plugin. It is responsible for making sure the plugin is downloaded and in
// the correct location before the bundling step begins.
type Plugin interface {
	// Latest fetches the Identifier for the latest available version that should be downloaded while respecting the
	// configuration of the user. For example, if the user specifies "0.1.7" as a version, this function will return
	// the identifier for that version of the plugin. The result is used to determine whether Plugin.Pull() should be
	// called. This identifier is then assumed to be the identifier of the newly updated plugin.
	Latest() (Identifier, error)
	// Pull will ensure all the necessary files for the plugin are downloaded, if this is needed. Extra checks can also
	// be done here to ensure that the plugin has been downloaded correctly. If an error is returned, the error will be
	// shown and the program will halt. This method is guaranteed to be executed before Plugin.Module().
	Pull() error
	// Module returns info about the go module of the plugin and the package that should be imported. It is used to
	// bundle the plugins.
	Module() Module
}

// Provider reads the plugin entry in the config data. If this provider type can successfully identify this plugin, it
// should be returned. If not, nil is returned, which indicates that the function cannot load this plugin. No actual
// loading should be done here.
type Provider = func(info map[string]any) (Plugin, error)

var providers []Provider

// RegisterProvider adds a new type of provider to the list of providers. The provider that was added last will be used
// first.
func RegisterProvider(p Provider) {
	providers = append(providers, p)
}

// ParseAll parses all plugins and tries to identify them. These plugins are then returned. This function does not take
// care of making sure plugins are downloaded.
func ParseAll(list []config.PluginInfo) ([]Plugin, error) {
	plugins := make([]Plugin, 0, len(list))
outerLoop:
	for num, info := range list {
		for _, provider := range providers {
			plugin, err := provider(info)
			if err != nil {
				return nil, pretty.Errorf("unable to parse plugin entry #%d: %v", num+1, err)
			}
			if plugin == nil {
				continue
			}
			plugins = append(plugins, plugin)
			continue outerLoop
		}
		return nil, pretty.Errorf("unable to parse plugin entry #%d.\n%v", num, info)
	}
	return plugins, nil
}
