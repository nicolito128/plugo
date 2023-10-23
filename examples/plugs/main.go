package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	router := plugo.New()

	router.Handle(plugo.MethodGet, "/", &HomePlug{})

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type HomePlug struct {
	*plugo.Plug
}

func (hm *HomePlug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Home"))
}
