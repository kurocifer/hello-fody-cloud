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
		fmt.Fprintf(w, "Hello future users of fody cloud. Soon you will be able to deploy your own projects with us :) %s\n", host)
	})

	http.HandleFunc("/eat", func(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 20; i++ {
		b := make([]byte, 10<<20) // 10 MiB
		for j := 0; j < len(b); j += 4096 {
			b[j] = 1 // touch every page so the memory is really used
		}
		hog = append(hog, b)
	}
	fmt.Fprintf(w, "holding %d MiB\n", len(hog)*10)
})

	fmt.Println("listening on port: ", port)
	http.ListenAndServe(":"+port, nil)
}