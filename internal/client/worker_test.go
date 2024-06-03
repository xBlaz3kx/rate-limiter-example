package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWorker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	clientName := "test"
	numRequests := atomic.Int32{}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		numRequests.Add(1)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, clientName, r.Form.Get("clientId"))
	}))
	defer svr.Close()

	// Create a worker and start it
	worker := newWorker(newHttpClient(clientName, svr.URL))
	worker.Start(ctx)

	<-ctx.Done()
	if numRequests.Load() <= 1 {
		t.Errorf("expected more than 1 request, got %d", numRequests.Load())
	}
}

func TestWorkerFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	clientName := "test"
	numRequests := atomic.Int32{}

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		numRequests.Add(1)

		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, clientName, r.Form.Get("clientId"))
	}))
	defer svr.Close()

	// Create a worker and start it
	worker := newWorker(newHttpClient(clientName, "http://invalid-url"))
	worker.Start(ctx)

	<-ctx.Done()
	reqs := numRequests.Load()
	if reqs > 0 {
		t.Errorf("expected no requests, got %d", reqs)
	}
}
