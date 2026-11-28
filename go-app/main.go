package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	mu   sync.Mutex
	hog  [][]byte
	once sync.Once
)

const maxMiB = 500

func eat() {
	for {
		mu.Lock()
		held := len(hog) * 10
		mu.Unlock()

		if held >= maxMiB {
			fmt.Println("reached safety cap, stopping at", held, "MiB")
			return
		}

		b := make([]byte, 10<<20) // 10 MiB
		for j := 0; j < len(b); j += 4096 {
			b[j] = 1 // touch every page so the memory is really used
		}

		mu.Lock()
		hog = append(hog, b)
		held = len(hog) * 10
		mu.Unlock()

		fmt.Println("holding", held, "MiB")
		time.Sleep(200 * time.Millisecond)
	}
}

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
		once.Do(func() { go eat() })
		fmt.Fprintln(w, "started eating memory, watch docker stats")
	})

	fmt.Println("listening on port: ", port)
	http.ListenAndServe(":"+port, nil)
}