package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	openai "github.com/sashabaranov/go-openai"

	"github.com/gorilla/mux"

	"plandex-server/handlers"
)

var client *openai.Client

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/prompt", handlers.PromptHandler).Methods("POST")
	r.HandleFunc("/summarize", handlers.SummarizeHandler).Methods("POST")

	// Get port from the environment variable or default to 8088
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("Plandex server is running on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
