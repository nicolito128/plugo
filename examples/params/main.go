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
	id, _ := conn.Param("id")
	return conn.String(http.StatusOK, fmt.Sprintf("ID: %s", id))
}

func home(conn plugo.Connection) error {
	return conn.String(http.StatusOK, "Home")
}
