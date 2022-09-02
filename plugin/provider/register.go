package provider

import "github.com/saddlemc/launcher/plugin"

func RegisterAll() {
	plugin.RegisterProvider(ModuleProvider)
	plugin.RegisterProvider(LocalProvider)
}
