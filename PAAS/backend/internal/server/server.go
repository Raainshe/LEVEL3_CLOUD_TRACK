package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"k8s.io/client-go/dynamic"

	"backend/internal/database"
	"backend/internal/kube"
)

type Server struct {
	port          int
	kubeClient    dynamic.Interface
	db            database.Service
	jwtSecret     string
	jwtTTLMinutes int
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	jwtTTLMinutes := 60
	if v := os.Getenv("JWT_TTL_MINUTES"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			jwtTTLMinutes = parsed
		}
	}

	kubeClient, err := kube.NewClient()
	if err != nil {
		log.Fatalf("failed to initialise kube client: %v", err)
	}
	srv := &Server{
		port:          port,
		kubeClient:    kubeClient,
		db:            database.New(),
		jwtSecret:     jwtSecret,
		jwtTTLMinutes: jwtTTLMinutes,
	}

	go srv.RunStatusPoller(context.Background())

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
