package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	// atomic int lets us safely access and update data over multiple goroutines
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// increment the fileserverHits counter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}

func (cfg *apiConfig) printHitsHandler(res http.ResponseWriter, req *http.Request) {
	// prints the number of fileserverHits
	num := cfg.fileserverHits.Load()
	txt := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", num)

	res.Header().Add("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(txt))
}

func (cfg *apiConfig) resetHitsHandler(res http.ResponseWriter, req *http.Request) {
	// resets the number of fileserverHits
	txt := strconv.AppendInt([]byte("Hits: "), 0, 10)

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(txt))
}

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func main() {

	// create the handler for the server
	cfg := apiConfig{}
	serve_mux := http.NewServeMux()
	serve_mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serve_mux.HandleFunc("GET /api/healthz", healthzHandler)
	serve_mux.HandleFunc("GET /admin/metrics", cfg.printHitsHandler)
	serve_mux.HandleFunc("POST /admin/reset", cfg.resetHitsHandler)

	// create the server
	server_struct := http.Server{}
	server_struct.Addr = ":8080"
	server_struct.Handler = serve_mux
	err := server_struct.ListenAndServe()
	if err != nil {
		fmt.Print(err)
	}
}
