package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host, _ := os.Hostname()
		fmt.Fprintln(w, "Ban... kai")
		fmt.Fprintf(w, "Hello future users of fody cloud. Soon you will be able to deploy your own projects with us :) %s\n", host)
	})

	fmt.Println("listening on port: ", port)
	http.ListenAndServe(":"+port, nil)
}