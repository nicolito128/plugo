package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	// Creating a new http router
	router := plugo.New()

	// Defines a new method GET for /
	router.Get("/", home)
	// Defines a new method GET for /hello/:world, here ':world' is a param value
	router.Get("/hello/:world", hello_world)

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func home(w http.ResponseWriter, r *http.Request) {
	conn := plugo.NewConnection(w, r)

	// Sending our first hello message
	conn.String(http.StatusOK, "Hello, Plugo World!")
}

func hello_world(w http.ResponseWriter, r *http.Request) {
	conn := plugo.NewConnection(w, r)

	// Catch the param if it exist
	s, ok := conn.Param("world")
	if !ok {
		s = "Anonymous"
	}

	conn.String(http.StatusOK, "Hello, %s", s)
}
