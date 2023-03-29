package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"jarnfast/signalman/pkg"
	"jarnfast/signalman/pkg/api"
	"jarnfast/signalman/pkg/cmdwrapper"
	"jarnfast/signalman/pkg/utl"

	"go.uber.org/zap"
)

func createLogger() *zap.SugaredLogger {
	var config zap.Config

	configFilename := os.Getenv("SIGNALMAN_LOG_CONFIGFILE")

	if configFilename != "" {
		configFile, err := os.Open(configFilename)
		if err != nil {
			panic(fmt.Sprintf("Unable to open config file: %v", err))

		}
		rawJSON, err := io.ReadAll(configFile)
		if err != nil {
			panic(fmt.Sprintf("Unable to read config file: %v", err))
		}

		if err := json.Unmarshal(rawJSON, &config); err != nil {
			panic(fmt.Sprintf("Unable to parse config file as JSON: %v", err))
		}

	} else {
		config = zap.NewProductionConfig()
		var err error
		config.Level, err = zap.ParseAtomicLevel(utl.GetenvDefault("SIGNALMAN_LOG_LEVEL", "info"))
		if err != nil {
			panic(err)
		}
	}

	logger, _ := config.Build()

	defer logger.Sync()

	sugar := logger.Sugar()
	return sugar
}

func main() {
	logger := createLogger()

	logger.Infof("Starting %s", pkg.VersionString())

	sigs := make(chan os.Signal, 10)

	// Relay incoming signals to process
	signal.Notify(sigs)

	a := api.NewApi(logger)
	// Relay signals received on HTTP
	a.Notify(sigs)
	a.ListenAndServe()

	c := cmdwrapper.NewCmdWrapper(logger)
	// Subscribe on the received signals
	c.Subscribe(sigs)
	c.Start()

	exitCode, err := c.Wait()

	close(sigs)

	if err != nil {
		exitCode := -1
		// try to get the original exit code
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		}
		logger.Warnf("Wrapped command finished with error: %v", err)
		os.Exit(exitCode)
	} else {
		logger.Info("Wrapped command finished without errors")
		os.Exit(exitCode)
	}
}
