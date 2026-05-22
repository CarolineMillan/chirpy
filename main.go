package main

import (
	"fmt"
	"net/http"
)

func main() {

	// create the handler for the server
	serve_mux := http.NewServeMux()
	serve_mux.Handle("/", http.FileServer(http.Dir(".")))

	// create the server
	server_struct := http.Server{}
	server_struct.Addr = ":8080"
	server_struct.Handler = serve_mux
	err := server_struct.ListenAndServe()
	if err != nil {
		fmt.Print(err)
	}
}
