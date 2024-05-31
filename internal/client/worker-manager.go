package client

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type WorkerManager struct {
	wg     *sync.WaitGroup
	logger *zap.Logger

	// ID of the client
	clientId string

	// URL to send requests to
	url string

	// Worker cancellation function
	cancelFunc context.CancelFunc
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(url, clientId string) *WorkerManager {
	return &WorkerManager{
		clientId:   clientId,
		url:        url,
		logger:     zap.L().Named(fmt.Sprintf("manager-%s", clientId)),
		wg:         &sync.WaitGroup{},
		cancelFunc: nil,
	}
}

// SpawnWorkers starts the worker manager with the provided number of workers. It blocks until all workers are done.
func (wm *WorkerManager) SpawnWorkers(ctx context.Context, numWorkers int) {
	wm.logger.Info("Starting worker manager", zap.Int("numWorkers", numWorkers))

	wm.wg.Add(numWorkers)

	// Create a context for the workers
	workerContext, cancel := context.WithCancel(context.Background())
	wm.cancelFunc = cancel

	// Spawn the workers
	for i := 0; i < numWorkers; i++ {
		workerHttpClient := newHttpClient(wm.clientId, wm.url)
		w := newWorker(workerHttpClient)

		// Run the worker asynchronously
		go func() {
			defer wm.wg.Done()
			w.Start(workerContext)
		}()
	}
}

func (wm *WorkerManager) Shutdown() {
	wm.logger.Info("Shutting down manager")

	// Send a cancellation signal to the workers
	wm.cancelFunc()

	// Wait for all workers to finish
	wm.wg.Wait()
}
