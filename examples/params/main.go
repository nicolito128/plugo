package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	router := plugo.New()

	router.Get("/", home)
	router.Get("/users/:id", users)

	fmt.Println("Server running at http://localhost:8080 - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func users(conn plugo.Connection) error {
	var id string
	if len(conn.URLParams()) > 0 {
		id = conn.URLParams()[0]
	}

	return conn.String(http.StatusOK, fmt.Sprintf("ID: %s", id))
}

func home(conn plugo.Connection) error {
	return conn.String(http.StatusOK, "Home")
}
