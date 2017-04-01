package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// hello world, the web server
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	http.HandleFunc("/", HelloServer)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("No PORT defined")
	}
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
