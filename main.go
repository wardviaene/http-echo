package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		text := os.Getenv("TEXT")
		if text == "" {
			fmt.Fprintf(w, "set env TEXT to display something")
		} else {
			fmt.Fprintf(w, text)
		}
	})

	fmt.Printf("Listening on port 8080\n")

	http.ListenAndServe(":8080", nil)
}
