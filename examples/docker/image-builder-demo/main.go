package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var port int

func main() {
	flag.IntVar(&port, "port", 8080, "http server port")
	flag.Parse()

	fn := func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "OK")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/helthz", fn)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: time.Second * 4,
		Handler:      mux,
	}

	idleConnsClosed := make(chan struct{})
	go func(s *http.Server) {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}(srv)

	go func(s *http.Server) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}(srv)

	log.Printf("server run at :%d", port)
	<-idleConnsClosed
	log.Print("server exit")
}
