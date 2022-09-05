package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/saddlemc/launcher/bundler"
	"github.com/saddlemc/launcher/config"
	"github.com/saddlemc/launcher/plugin"
	"github.com/saddlemc/launcher/plugin/provider"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	// The code currently found in the program, and especially this main function, might currently be a bit messy. This
	// is something that is planned to be worked on to improve it in the future.

	var logger *zerolog.Logger
	{
		l := zerolog.New(os.Stdout).
			Level(zerolog.DebugLevel).
			Output(zerolog.ConsoleWriter{
				Out:          os.Stdout,
				PartsExclude: []string{zerolog.TimestampFieldName},
			})
		logger = &l
	}

	// Get all flags. They may override some settings in the configuration.
	flagOut := flag.String("out", "",
		"Specifies an output file name for the server binary.",
	)
	flagRecompile := flag.Bool("recompile", false,
		"If set to true, the server will always be recompiled.",
	)
	flag.Parse()

	logger.Debug().Msgf("Reading saddle.toml...")
	cfg := config.GetOrMakeConfig(logger, "saddle.toml")
	if *flagOut != "" {
		cfg.Bundler.Path = *flagOut
	}

	// On Windows, a '.exe' is required after the server executable.
	outFile, err := filepath.Abs(cfg.Bundler.Path)
	if err != nil {
		logger.Panic().Msgf("Unable to get current working directory.")
	}
	if runtime.GOOS == "windows" {
		if strings.ToLower(filepath.Ext(outFile)) != ".exe" {
			outFile += ".exe"
		}
	}

	logger.Info().Msgf("Checking for updates...")

	logger.Debug().Msgf("Parsing plugins...")
	provider.RegisterAll()
	plugins, err := plugin.ParseAll(cfg.Plugin)
	if err != nil {
		logger.Fatal().Msgf("Error trying to parse plugins: %v", err)
	}

	logger.Debug().Msgf("Reading saddle.lock...")
	// Get the current lockfile and also make a new lockfile. After checking plugin versions, the two will be compared
	// to see if the already present executable is outdated.
	needsRebuilding := false
	lock, ok := config.GetLock(logger, "saddle.lock")
	if !ok {
		// If the lockfile could not successfully be loaded we rebuild the server regardless.
		needsRebuilding = true
	}
	newLock := config.LockFile{
		Version:   config.LockVersion,
		Api:       cfg.Server.Api,
		Dragonfly: cfg.Server.Dragonfly,
		Plugins:   map[string]string{},
	}
	pluginModules := make([]plugin.Module, 0, len(plugins))
	for num, pl := range plugins {
		latest, err := pl.Latest()
		if err != nil {
			logger.Fatal().Msgf("Error trying to fetch latest version for plugin entry #%d: %v", num, err)
		}
		if x, ok := lock.Plugins[latest.Module]; !ok || x != latest.Checksum {
			needsRebuilding = true
			err = pl.Pull()
			if err != nil {
				logger.Fatal().Msgf("Error trying to update plugin entry #%d: %v", num, err)
			}
		}

		newLock.Plugins[latest.Module] = latest.Checksum
		pluginModules = append(pluginModules, pl.Module())
	}

	// Check if the file exist and can be accessed. If the file does not exist, we always have to rebuild the server.
	if _, err = os.Stat(outFile); os.IsNotExist(err) {
		logger.Debug().Msgf("No server binary detected, force rebuilding server.")
		needsRebuilding = true
	} else if err != nil {
		logger.Error().Msgf("Unable to access output location: %s", err)
	}
	// Rebuilt the server is there was an update or if the '--recompile' flag was passed.
	if needsRebuilding || *flagRecompile {
		logger.Info().Msgf("Rebuilding server...")
		buildStart := time.Now()
		// Create a temporary directory to build the server in.
		temp, err := os.MkdirTemp("", "saddle_bundler_*")
		if err != nil {
			logger.Fatal().Msgf("Could not create temporary directory: %v", err)
		}
		logger.Debug().Msgf("Created temporary directory '%s'.", temp)
		// Be sure to remove the temporary directory after creating it.
		defer os.RemoveAll(temp)

		logger.Debug().Msgf("Bundling plugins...")
		err = bundler.Bundle(makeBundleConfig(logger, cfg, temp, pluginModules))
		if err != nil {
			logger.Fatal().Msgf("Could not bundle plugins: %v", err)
		}

		{
			logger.Debug().Msgf("Compiling server...")
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
		logger.Info().Msgf("Done! Finished building in %.3f seconds.", time.Now().Sub(buildStart).Seconds())

		// The server has been built successfully. Now store the build information as the new lock file.
		logger.Debug().Msgf("Writing saddle.lock...")
		data, err := json.Marshal(newLock)
		if err != nil {
			logger.Fatal().Msgf("Could not encode saddle.lock: %v", err)
		}

		f, err := os.Create("saddle.lock")
		if err != nil {
			logger.Fatal().Msgf("Could not open saddle.lock: %v", err)
		}
		defer f.Close()

		_, err = f.Write(data)
		if err != nil {
			logger.Fatal().Msgf("Could not write saddle.lock: %v", err)
		}
	}

	// Run the server.
	{
		cmd := exec.Command(outFile)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = filepath.Dir(outFile)
		err = cmd.Start()
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}

		// Wait for the program to end, and report any error that might have occurred.
		shutdown := make(chan error)
		go func() {
			shutdown <- cmd.Wait()
		}()

		// When ctrl+c is pressed, make sure to wait for the server to close so the user can see all the output. We wait
		// for up to 10 seconds for this shutdown to be successful before the server is forcibly killed.
		timeout := make(chan struct{})
		go func() {
			c := make(chan os.Signal, 2)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			<-c

			time.Sleep(time.Second * 10)
			close(timeout)
		}()

		// Either wait for the graceful shutdown or for the server to time out.
		select {
		case err = <-shutdown:
			if _, ok := err.(*exec.ExitError); err != nil && ok {
				fmt.Println("")
				logger.Fatal().Msgf(
					"A fatal error caused the server to shut down unexpectedly. When reporting this error, " +
						"please include the entire error message above, as well as your saddle.lock file at the time " +
						"of the error. Consider trying to find the probable cause before opening an issue for the " +
						"correct plugin.",
				)
			} else if err != nil {
				logger.Fatal().Msgf("Error shutting down server: %s", err)
			}
		case _, _ = <-timeout:
			logger.Error().Msgf("Server shutdown took to long, killing server...")
			err := cmd.Process.Kill()
			if err != nil {
				logger.Fatal().Msgf("Error killing server: %s", err)
			}
		}
	}
}

func makeBundleConfig(logger *zerolog.Logger, cfg *config.Config, path string, pluginModules []plugin.Module) bundler.Settings {
	// absIfLocal is a helper function used in this function. It makes the path absolute if not empty.
	// todo: providing a local dragonfly or saddle location does not currently trigger a rebuild
	absIfLocal := func(s string) string {
		if s != "" {
			var err error
			s, err = filepath.Abs(s)
			if err != nil {
				logger.Panic().Msgf("Unable to get current working directory.")
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
	}
}
