package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/xBlaz3kx/rate-limiter-example/internal/client"
	"go.uber.org/zap"
)

var (
	numClients = 2
	numWorkers = 4
)

var rootCmd = &cobra.Command{
	Use:       "client",
	ValidArgs: []string{"url"},
	Version:   "0.1.0",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a URL argument")
		}

		// Check if the URL is valid
		_, err := url.Parse(args[0])
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

		logger := zap.L()
		logger.Info("Starting the clients")

		// Get the URL as the first argument
		url := args[0]

		// todo get and validate number of clients and workers

		workerManagers := make([]*client.WorkerManager, 0)

		for i := 0; i < numClients; i++ {
			clientId := strconv.Itoa(i + 1)

			workerManager := client.NewWorkerManager(url, clientId)
			workerManager.SpawnWorkers(ctx, numWorkers)

			workerManagers = append(workerManagers, workerManager)
		}

		// Receive a shutdown signal
		<-ctx.Done()
		cancel()

		// Shutdown all worker managers
		for _, wm := range workerManagers {
			wm.Shutdown()
		}
	},
}

func main() {
	cobra.OnInitialize(setupGlobalLogger)

	if err := rootCmd.Execute(); err != nil {
		zap.L().Fatal("Unable to run", zap.Error(err))
	}
}

func setupGlobalLogger() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}
