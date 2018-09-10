package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		latency := os.Getenv("LATENCY")
		if latency != "" {
			i, err := strconv.ParseInt(latency, 10, 64)
			if err != nil {
				fmt.Fprintf(w, "Env LATENCY needs to be a number")
				return
			}
			time.Sleep(time.Duration(i) * time.Second)
		}
		text := os.Getenv("TEXT")
		if text == "" {
			fmt.Fprintf(w, "set env TEXT to display something")
			return
		}
		next := os.Getenv("NEXT")
		if next == "" {
			fmt.Fprintf(w, "%s", text)
		} else {
			resp, err := http.Get("http://" + next + "/")
			if err != nil {
				fmt.Fprintf(w, "Couldn't connect to http://%s/", next)
				fmt.Printf("Error: %s", err)
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(w, "%s %s\n", text, body)
		}
	})

	fmt.Printf("Listening on port 8080\n")

	http.ListenAndServe(":8080", nil)
}
