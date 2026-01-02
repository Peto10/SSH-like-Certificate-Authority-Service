package api

import (
	"net/http"
	"log/slog"
)

type Controller struct {
	Log *slog.Logger
}

func NewController(logger *slog.Logger) *Controller {
	return &Controller{Log: logger}
}

func (c *Controller) Sign(w http.ResponseWriter, r *http.Request) {
	// TODO parse Authorisation header and validate request
	c.Log.Info("Received /sign request", "headers", r.Header)
	w.WriteHeader(http.StatusNotImplemented)
}
