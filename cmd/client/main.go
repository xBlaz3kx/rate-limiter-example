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
	numClients int
	numWorkers int
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

		if numClients < 1 {
			logger.Fatal("Number of clients must be greater than 0")
		}

		if numWorkers < 1 {
			logger.Fatal("Number of workers must be greater than 0")
		}

		clientManagers := []*client.Manager{}

		for i := 0; i < numClients; i++ {
			clientId := strconv.Itoa(i + 1)

			workerManager := client.NewManager(url, clientId)
			workerManager.SpawnClients(numWorkers)

			clientManagers = append(clientManagers, workerManager)
		}

		// Receive a shutdown signal
		<-ctx.Done()
		cancel()

		// Shutdown all worker managers
		for _, wm := range clientManagers {
			wm.Shutdown()
		}
	},
}

func main() {
	cobra.OnInitialize(setupGlobalLogger)

	rootCmd.Flags().IntVar(&numClients, "clients", 1, "Number of clients to spawn")
	rootCmd.Flags().IntVar(&numWorkers, "workers", 1, "Number of workers per client")

	if err := rootCmd.Execute(); err != nil {
		zap.L().Fatal("Unable to run", zap.Error(err))
	}
}

func setupGlobalLogger() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
}
