package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nicolito128/plugo"
)

func main() {
	// Creating a new http router
	router := plugo.New()

	router.Use(logger)

	router.Get("/", home)

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// logger middleware
func logger(fail *error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf(
			"Logger :: %s |> METHOD: %s |> PATH: %s |> HOST: %s \n",
			time.Now().Format(time.UnixDate),
			r.Method,
			r.URL.Path,
			r.Host,
		)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	conn := plugo.NewConnection(w, r)

	// Sending our first hello message
	conn.String(http.StatusOK, "Hello, Plugo World!")
}
