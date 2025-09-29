package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func main() {
	start := time.Now()

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Println("Listening on", ln.Addr().String())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	go func() {
		if err := http.Serve(ln, nil); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("Server started in %s\n", time.Since(start))

	ip := fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	)

	fmt.Println("Random IPv4:", ip)

	select {}
}
