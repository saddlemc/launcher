package plugin

// Module contains data about the plugin that should be bundled and where it can be found. This is used to generate
// a go.mod file and a main.go file for the server executable.
type Module struct {
	// Module is the name and version of the go module that the plugin is in. Generally, this will be formatted as
	// "site.com/author/repo". This will be added to the go.mod file.
	Module string
	// Version is the version string used in the require directive for the package. It is not use anywhere else, and can
	// be left empty if the plugin does not have a version. An empty version string will be treated as "v0.0.0"
	Version string
	// Replace allows the plugin to optionally add a replace directive to the go.mod file for this plugin. This is
	// usually a path to a local go module.
	Replace string
	// Import returns the package that should be imported for side effects in the server. This is full name of the
	// package (equivalent to how it would be imported) where the plugin is registered to saddle, such as
	// "site.com/author/repo/plugin".
	Import string
}
