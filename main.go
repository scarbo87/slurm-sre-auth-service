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

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	address    string
	authHeader string
	authToken  string

	okResponse  = []byte(`{}`)
	errResponse = []byte(`{"error":"Authorization failed"}`)
)

func init() {
	address = os.Getenv("AUTH_SERVICE_ADDRESS")
	if address == "" {
		address = ":2121"
	}

	authHeader = os.Getenv("AUTH_SERVICE_HEADER")
	if authHeader == "" {
		authHeader = "X-Slurm-Source-Provider"
	}

	authToken = os.Getenv("AUTH_SERVICE_TOKEN")
	if authToken == "" {
		authToken = "belgorod"
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-ch
		cancel()
	}()

	log.Printf("Run the auth-service by address %s\n", address)
	server := &http.Server{Addr: address, Handler: getHttpHandler()}
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("An error %v occurred while running the auth-service\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown the auth-service")

	ctx, cancel = context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("An error %v occurred while closing the auth-service\n", err)
	}
}

func getHttpHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)
	mux.HandleFunc("/health", func(rw http.ResponseWriter, request *http.Request) {
		_, _ = rw.Write(okResponse)
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(authHeader)
		if token != authToken {
			log.Printf("Authorization failed, token: %s\n", token)

			rw.Header().Set("Content-Type", "application/json; charset=utf-8")
			rw.WriteHeader(http.StatusUnauthorized)
			_, _ = rw.Write(errResponse)
			return
		}

		operationId := uuid.New().String()
		log.Printf("Authorization successful, operation-id: %s\n", operationId)

		rw.Header().Set("X-Auth-Operation-Id", operationId)
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = rw.Write(okResponse)
	})
	return mux
}
