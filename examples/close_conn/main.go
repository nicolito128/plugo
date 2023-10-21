package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	router := plugo.New()

	router.Prev(func() plugo.HandlerFunc {
		return func(conn plugo.Connection) error {
			fmt.Println("Closing connection to future requests...")
			conn.Done()

			return conn.String(http.StatusOK, "Connection was closed!")
		}
	})

	router.Get("/", hello)

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func hello(conn plugo.Connection) error {
	return conn.String(http.StatusOK, "Hello!")
}
