package main

import (
	"fmt"
	"net/http"
)

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func main() {

	// create the handler for the server
	serve_mux := http.NewServeMux()
	serve_mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	serve_mux.HandleFunc("/healthz", healthzHandler)

	// create the server
	server_struct := http.Server{}
	server_struct.Addr = ":8080"
	server_struct.Handler = serve_mux
	err := server_struct.ListenAndServe()
	if err != nil {
		fmt.Print(err)
	}
}
