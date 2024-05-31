package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xBlaz3kx/rate-limiter-example/internal/server/rate-limiter"
)

type Handler struct {
	limiter *rate_limiter.SlidingWindowRateLimiter
}

func NewHandler(limiter *rate_limiter.SlidingWindowRateLimiter) *Handler {
	return &Handler{
		limiter: limiter,
	}
}

func (h *Handler) HandleRequest(ctx *gin.Context) {
	clientId, isFound := ctx.GetQuery("clientId")
	if !isFound || clientId == "" {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "clientId is required"})
		return
	}

	if h.limiter.IsLimited(clientId) {
		ctx.JSON(http.StatusTooManyRequests, errorResponse{Error: "rate limit exceeded"})
	} else {
		ctx.JSON(http.StatusNoContent, nil)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}
