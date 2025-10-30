package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/temirov/pinguin/clients/cli/internal/command"
	cliConfig "github.com/temirov/pinguin/clients/cli/internal/config"
	"github.com/temirov/pinguin/pkg/client"
	"github.com/temirov/pinguin/pkg/logging"
)

func main() {
	v := viper.New()
	cfg, err := cliConfig.Load(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	settings, err := client.NewSettings(
		cfg.ServerAddress(),
		cfg.AuthToken(),
		cfg.ConnectionTimeoutSeconds(),
		cfg.OperationTimeoutSeconds(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := logging.NewLogger(cfg.LogLevel())

	notificationClient, err := client.NewNotificationClient(logger, settings)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer notificationClient.Close()

	root := command.NewRootCommand(command.Dependencies{
		Sender:           notificationClient,
		OperationTimeout: cfg.OperationTimeout(),
		Output:           os.Stdout,
	})
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	if execErr := root.Execute(); execErr != nil {
		fmt.Fprintln(os.Stderr, execErr)
		os.Exit(1)
	}
}
