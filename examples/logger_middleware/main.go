package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	router := plugo.New()

	router.Use(logger)

	router.Get("/", func(conn plugo.Connection) error {
		return conn.String(http.StatusOK, "Hello, Plugo World!")
	})

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func logger() plugo.HandlerFunc {
	return func(conn plugo.Connection) error {
		req := conn.Request()
		res := conn.Response()
		fmt.Println(fmt.Sprintf("|> Method %s | Status %d | Path: %s", req.Method, res.Status(), req.RequestURI))

		return nil
	}
}
