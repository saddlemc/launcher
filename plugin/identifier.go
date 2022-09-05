package plugin

import (
	"github.com/rogpeppe/go-internal/dirhash"
	"github.com/rogpeppe/go-internal/modfile"
	"os"
	path2 "path"
)

// Identifier contains general info about a plugin that uniquely identifies it. It is used to store which plugins were
// already bundled with the server executable.
type Identifier struct {
	// Module is the name of the go module that the plugin is in. Generally, this will be formatted as
	// "site.com/author/repo". This will be added to the go.mod file.
	Module string
	// Checksum is a unique string of data that identifies one specific version of a plugin. It is used to determine if
	// the plugin has been updated since the last time the server has been compiled.
	// Any algorithm may be used, as long as it is deterministic and unique for a specific version of a plugin's source
	// code and the provider that was used to get the plugin. For local plugins, it may be the checksum of the entire
	// directory where the plugin's go.mod file is in. For remote git repositories, the commit hash works.
	Checksum string
}

// ParseIdentifier parses Identifier for a local plugin. It accepts the local path of the plugin as a first parameter.
func ParseIdentifier(path string) (Identifier, error) {
	f, err := os.ReadFile(path2.Join(path, "go.mod"))
	if err != nil {
		return Identifier{}, err
	}
	parse, err := modfile.Parse(path, f, nil)
	if err != nil {
		return Identifier{}, err
	}
	hash, err := dirhash.HashDir(path, "", dirhash.Hash1)
	if err != nil {
		return Identifier{}, err
	}
	return Identifier{
		Module:   parse.Module.Mod.Path,
		Checksum: hash,
	}, nil
}
