package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/config"
	"github.com/xBlaz3kx/rate-limiter-example/internal/server/api"
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
		ginHandler := api.NewHandler(slidingWindowLimiter)

		router := gin.New()
		// Setup logging and recovery middleware
		router.Use(ginzap.Ginzap(logger, time.RFC3339, true), ginzap.RecoveryWithZap(logger, true))
		_ = healthcheck.New(router, config.DefaultConfig(), nil)

		router.GET("", ginHandler.HandleRequest)

		srv := &http.Server{
			Addr:    ":80",
			Handler: router.Handler(),
		}

		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("Failed to listen and serve", zap.Error(err))
			}
		}()

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 5 seconds.
		<-quit
		logger.Info("Shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}

		<-ctx.Done()
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
