package config

import (
	_ "embed"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog"
	"os"
)

//go:embed default_config.toml
var defaultConfig []byte

type PluginInfo = map[string]any

type Config struct {
	Bundler struct {
		Debug bool   `toml:"debug-log"`
		Path  string `toml:"server-path"`
	}

	Server struct {
		Api string `toml:"api"`
		// ApiReplace is an optional argument that allows a local saddle API version to be used instead of one pulled
		// from GitHub.
		ApiReplace string `toml:"replace-api"`
		Dragonfly  string `toml:"dragonfly"`
		// DragonflyReplace is an optional argument that allows a local dragonfly version to be used instead of one pulled
		// from GitHub.
		DragonflyReplace string `toml:"replace-dragonfly"`
	} `toml:"server"`

	Plugin []PluginInfo `toml:"plugin"`
}

// GetOrMakeConfig tries to load the config file, and if it does not exist the default config file will be created and
// loaded.
func GetOrMakeConfig(log *zerolog.Logger, path string) *Config {
	cfg := &Config{}
	cfgData, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Info().Msgf("Config file does not exist, creating default config...")
		// Create the now config.toml file.
		file, err := os.Create(path)
		if err != nil {
			log.Fatal().Msgf("Error trying to create saddle.toml file: %v", err)
		}
		defer file.Close()
		_, err = file.Write(defaultConfig)
		if err != nil {
			log.Fatal().Msgf("Error trying to create saddle.toml file: %v", err)
		}

		// Make sure the new data is parsed.
		cfgData = defaultConfig
	} else if err != nil {
		log.Fatal().Msgf("Error trying to open saddle.toml file: %v", err)
	}

	err = toml.Unmarshal(cfgData, cfg)
	if err != nil {
		log.Fatal().Msgf("Error trying to parse saddle.toml file: %v", err)
	}
	return cfg
}
