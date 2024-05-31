package client

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

type worker struct {
	client client
	logger *zap.Logger
}

// newWorker creates a new worker for a client.
func newWorker(client client) *worker {
	return &worker{
		logger: zap.L().Named("worker"),
		client: client,
	}
}

// Start starts sending the requests using the provided client with a random delay.
func (w *worker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping worker")
			return
		default:
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

			// Send the request and delay for a random amount of time
			err := w.client.SendRequest(ctx)
			if err != nil {
				w.logger.Error("Failed to send request", zap.Error(err))
			}

			cancel()
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}
}
