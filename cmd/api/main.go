package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	log.Info("starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error("Server failed", "err", err)
		os.Exit(1)
	}
}
