package config

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
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

// GetLock returns the current lockfile. If it does not exist, or if the lockfile is of a previous version, the data is
// discarded and an empty lockfile is returned.
func GetLock(log logrus.FieldLogger, path string) *LockFile {
	lf := &LockFile{
		Version: LockVersion,
		Plugins: map[string]string{},
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return lf
	} else if err != nil {
		log.Fatalf("Error trying to open saddle.lock: %v", err)
	}

	err = json.Unmarshal(data, lf)
	if err != nil {
		// The saddle.lock data is only used to check if the server needs recompiling. In the event that the file could
		// not be parsed (it may be outdated), an empty saddle.lock is returned instead.
		log.Errorf("Error trying to parse saddle.lock: %v. Using an empty saddle.lock file.", err)
		lf = &LockFile{
			Version: LockVersion,
			Plugins: map[string]string{},
		}
	}
	if lf.Version > LockVersion {
		// Do not override newer versions of the lockfile. We don't know if this may contain any important data in the
		// future
		log.Fatalf("Unknown lockfile version %d.", lf.Version)
	} else if lf.Version < LockVersion {
		// Older versions of the lockfile can be safely discarded. In this case we make
		lf = &LockFile{
			Version: LockVersion,
			Plugins: map[string]string{},
		}
	}
	return lf
}
