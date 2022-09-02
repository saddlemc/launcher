package bundler

// Settings defines what should be bundled into a program and how it should be bundled.
type Settings struct {
	// Path is the path where the program should be created.
	Path string
	// Modules is a list of all the modules that should be added to the go.mod file.
	Modules []Module
	// Imports lists all packages that should be imported in the main.go file.
	Imports []Import
	// Run is a fragment of code that should be added to the main function.
	Run string
	// Go is the minimum go version that should be listed in the go.mod file.
	Go string
}

// Module represents a go module that is to be added to the go.mod file.
type Module struct {
	// Name is the name of the module, usually 'site.com/user/repo'. It is formatted like it would be in the go.mod file
	// or the go get command.
	Name string
	// Version is the version string of the module, which is usually a git ref. It accepts anything the go get command
	// does. If not provided, it will default to 'v0.0.0'.
	Version string
	// Replace allows the package to have an optional replace directive. It is generally replaced with a local path.
	Replace string
}

// Import allows for package imports to be specified in a file.
type Import struct {
	// Package is the name of the package to be imported.
	Package string
	// Alias is the import alias that the package should get. If not provided, this will default to an underscore '_'.
	Alias string
}
