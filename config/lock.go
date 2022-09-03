package config

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"os"
)

const LockVersion = 1

// LockFile contains information about the currently existing server binaries. It is used to determine whether it is
// up-to-date with the latest configuration.
type LockFile struct {
	// Version is the version the lockfile was made in.
	Version uint
	// Api is the version of the saddle api.
	Api string
	// Dragonfly is the dragonfly version used.
	Dragonfly string
	// Plugins is a map which contains all the plugin checksums. The keys are the plugin module names.
	Plugins map[string]string
}

// GetLock returns the current lockfile. If it does not exist, or if the lockfile is of a previous version, an empty
// lockfile and false will be returned.
func GetLock(log *zerolog.Logger, path string) (LockFile, bool) {
	lf := LockFile{
		Version: LockVersion,
		Plugins: map[string]string{},
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return lf, false
	} else if err != nil {
		log.Fatal().Msgf("Error trying to open saddle.lock: %v", err)
	}

	err = json.Unmarshal(data, &lf)
	if err != nil {
		// The saddle.lock data is only used to check if the server needs recompiling. In the event that the file could
		// not be parsed (it may be outdated), an empty saddle.lock is returned instead.
		log.Error().Msgf("Error trying to parse saddle.lock: %v. Using an empty saddle.lock file.", err)
		return LockFile{
			Version: LockVersion,
			Plugins: map[string]string{},
		}, false
	}
	if lf.Version > LockVersion {
		// Do not override newer versions of the lockfile. We don't know if this may contain any important data in the
		// future
		log.Fatal().Msgf("Unknown lockfile version %d.", lf.Version)
	} else if lf.Version < LockVersion {
		// Older versions of the lockfile can be safely discarded. In this case we make
		return LockFile{
			Version: LockVersion,
			Plugins: map[string]string{},
		}, false
	}
	return lf, true
}
