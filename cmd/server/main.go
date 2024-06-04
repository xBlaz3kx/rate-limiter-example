package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	http2 "github.com/xBlaz3kx/rate-limiter-example/internal/server/api/http"
	ratelimiter "github.com/xBlaz3kx/rate-limiter-example/internal/server/rate-limiter"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		logger := zap.L()
		logger.Info("Starting the server")

		// Set up the rate limiter
		slidingWindowLimiter := ratelimiter.NewSlidingWindowRateLimiter()

		// Set up the handler
		ginHandler := http2.NewHandler(slidingWindowLimiter)
		// Create a new HTTP server
		server := http2.NewServer(":80", logger)
		server.Router.GET("", ginHandler.HandleRequest)

		// Start the server
		server.Start()

		// Wait for interrupt signal to gracefully shutdown the server
		<-quit
		logger.Info("Shutting down server")

		err := server.Shutdown()
		if err != nil {
			logger.Fatal("Failed to shutdown server", zap.Error(err))
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
