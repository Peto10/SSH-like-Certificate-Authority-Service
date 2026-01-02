package main

import (
	"log/slog"
	"os"
	"fmt"

	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/api"
	"github.com/Peto10/SSH-like-Certificate-Authority-Service/internal/server"
	"github.com/joho/godotenv"
)

const (
	defaultServerHostName = ":8080"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	c := api.NewController(logger)
	godotenv.Load("../.env")

	c.Log.Info("service starting", "URL", fmt.Sprintf("http://localhost%s", defaultServerHostName))

	s := server.NewServer(c, defaultServerHostName)

	err := s.ListenAndServe()
	if err != nil {
		c.Log.Error("Error with listening and serving port", "error", err)
		os.Exit(1)
	}
}
