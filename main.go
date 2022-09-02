package main

import (
	"encoding/json"
	"flag"
	"github.com/saddlemc/launcher/bundler"
	"github.com/saddlemc/launcher/config"
	"github.com/saddlemc/launcher/plugin"
	"github.com/saddlemc/launcher/plugin/provider"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	// The code currently found in the program, and especially this main function, might currently be a bit messy. This
	// is something that is planned to be worked on to improve it in the future.

	log := logrus.StandardLogger()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.TextFormatter{ForceColors: true, DisableTimestamp: true}

	// Get all flags. They may override some settings in the configuration.
	flagOut := flag.String("out", "",
		"Specifies an output file name for the server binary.",
	)
	flagRecompile := flag.Bool("recompile", false,
		"If set to true, the server will always be recompiled.",
	)
	flag.Parse()

	log.Debugf("Reading saddle.toml...")
	cfg := config.GetOrMakeConfig(log, "saddle.toml")
	if *flagOut != "" {
		cfg.Bundler.Path = *flagOut
	}

	// On Windows, a '.exe' is required after the server executable.
	outFile, err := filepath.Abs(cfg.Bundler.Path)
	if err != nil {
		log.Panicf("Unable to get current working directory.")
	}
	if runtime.GOOS == "windows" {
		if strings.ToLower(filepath.Ext(outFile)) != ".exe" {
			outFile += ".exe"
		}
	}

	log.Infof("Checking for updates...")

	log.Debugf("Parsing plugins...")
	provider.RegisterAll()
	plugins, err := plugin.ParseAll(cfg.Plugin)
	if err != nil {
		log.Fatalf("Error trying to parse plugins: %v", err)
	}

	log.Debugf("Reading saddle.lock...")
	// Get the current lockfile and also make a new lockfile. After checking plugin versions, the two will be compared
	// to see if the already present
	lock := config.GetLock(log, "saddle.lock")
	newLock := config.LockFile{
		Version:   config.LockVersion,
		Api:       cfg.Server.Api,
		Dragonfly: cfg.Server.Dragonfly,
		Plugins:   map[string]string{},
	}
	needsRebuilding := false
	pluginModules := make([]plugin.Module, 0, len(plugins))
	for num, pl := range plugins {
		latest, err := pl.Latest()
		if err != nil {
			log.Fatalf("Error trying to fetch latest version for plugin entry #%d: %v", num, err)
		}
		if x, ok := lock.Plugins[latest.Module]; !ok || x != latest.Checksum {
			needsRebuilding = true
			err = pl.Pull()
			if err != nil {
				log.Fatalf("Error trying to update plugin entry #%d: %v", num, err)
			}
		}

		newLock.Plugins[latest.Module] = latest.Checksum
		pluginModules = append(pluginModules, pl.Module())
	}

	// Rebuilt the server is there was an update or if the '--recompile' flag was passed.
	if needsRebuilding || *flagRecompile {
		log.Infoln("Rebuilding server...")
		buildStart := time.Now()
		// Create a temporary directory to build the server in.
		temp, err := os.MkdirTemp("", "saddle_bundler_*")
		if err != nil {
			log.Fatalf("Could not create temporary directory: %v", err)
		}
		log.Debugf("Created temporary directory '%s'.", temp)
		// Be sure to remove the temporary directory after creating it.
		defer os.RemoveAll(temp)

		log.Debugln("Bundling plugins...")
		err = bundler.Bundle(makeBundleConfig(cfg, temp, pluginModules))
		if err != nil {
			log.Fatalf("Could not bundle plugins: %v", err)
		}

		{
			log.Debugln("Compiling server...")
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = temp
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				panic(err)
			}

			cmd = exec.Command("go", "build", "-o", outFile)
			cmd.Dir = temp
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				panic(err)
			}
		}
		log.Infof("Done! Finished building in %.3f seconds.", time.Now().Sub(buildStart).Seconds())

		// The server has been built successfully. Now store the build information as the new lock file.
		log.Debugf("Writing saddle.lock...")
		data, err := json.Marshal(newLock)
		if err != nil {
			log.Fatalf("Could not encode saddle.lock: %v", err)
		}

		f, err := os.Create("saddle.lock")
		if err != nil {
			log.Fatalf("Could not open saddle.lock: %v", err)
		}
		defer f.Close()

		_, err = f.Write(data)
		if err != nil {
			log.Fatalf("Could not write saddle.lock: %v", err)
		}
	}

	// Run the server.
	{
		cmd := exec.Command(outFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = filepath.Dir(outFile)
		err = cmd.Run()
		if err != nil {
			panic(err)
		}
	}
}

func makeBundleConfig(cfg *config.Config, path string, pluginModules []plugin.Module) bundler.Settings {
	// absIfLocal is a helper function used in this function. It makes the path absolute if not empty.
	// todo: providing a local dragonfly or saddle location does not currently trigger a rebuild
	absIfLocal := func(s string) string {
		if s != "" {
			var err error
			s, err = filepath.Abs(s)
			if err != nil {
				log.Panicf("Unable to get current working directory.")
			}
		}
		return s
	}

	dfReplace := absIfLocal(cfg.Server.DragonflyReplace)
	apiReplace := absIfLocal(cfg.Server.ApiReplace)
	// Insert dragonfly and saddle into the bundler configuration.
	var (
		modules = append(make([]bundler.Module, 0, len(pluginModules)+2),
			bundler.Module{
				Name:    "github.com/df-mc/dragonfly",
				Version: cfg.Server.Dragonfly,
				Replace: dfReplace,
			},
			bundler.Module{
				Name:    "github.com/saddlemc/saddle",
				Version: cfg.Server.Api,
				Replace: apiReplace,
			},
		)
		imports = append(make([]bundler.Import, 0, len(pluginModules)+1),
			bundler.Import{
				Package: "github.com/saddlemc/saddle",
				Alias:   ".",
			},
		)
	)
	// Convert all the plugin information to information that the bundler accepts.
	for _, pl := range pluginModules {
		modules = append(modules, bundler.Module{
			Name:    pl.Module,
			Version: pl.Version,
			Replace: pl.Replace,
		})
		imports = append(imports, bundler.Import{
			Package: pl.Import,
			Alias:   "_",
		})
	}
	return bundler.Settings{
		Path:    path,
		Modules: modules,
		Imports: imports,
		Run:     "Run()",
		Go:      strings.TrimPrefix(runtime.Version(), "go"),
	}
}
