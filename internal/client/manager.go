package client

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type Manager struct {
	wg     *sync.WaitGroup
	logger *zap.Logger

	// ID of the client
	clientId string

	// URL to send requests to
	url string

	// Worker cancellation function
	cancelFunc context.CancelFunc
}

// NewManager creates a new client manager
func NewManager(url, clientId string) *Manager {
	return &Manager{
		clientId:   clientId,
		url:        url,
		logger:     zap.L().Named(fmt.Sprintf("manager-%s", clientId)),
		wg:         &sync.WaitGroup{},
		cancelFunc: nil,
	}
}

// SpawnClients starts the worker manager with the provided number of workers. Each worker runs in its own goroutine.
func (wm *Manager) SpawnClients(num int) {
	wm.logger.Info("Starting worker manager", zap.Int("num", num))

	wm.wg.Add(num)

	// Create a context for the workers
	workerContext, cancel := context.WithCancel(context.Background())
	wm.cancelFunc = cancel

	// Spawn the workers
	for i := 0; i < num; i++ {
		workerHttpClient := newHttpClient(wm.clientId, wm.url)
		w := newWorker(workerHttpClient)

		// Run the worker asynchronously
		go func() {
			defer wm.wg.Done()
			w.Start(workerContext)
		}()
	}
}

// Shutdown cancels all workers' context and waits for them to finish executing.
func (wm *Manager) Shutdown() {
	wm.logger.Info("Shutting down manager")

	// Send a cancellation signal to the workers
	wm.cancelFunc()

	// Wait for all workers to finish
	wm.wg.Wait()
}
