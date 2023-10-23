# Plugo
Composable and minimalistic HTTP router for building Go backend applications inspired by [Plug](https://hex.pm/packages/plug).

## Quick start
Require `go 1.21.0` or higher.

### Installation

    go get github.com/nicolito128/plugo

### Example

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/nicolito128/plugo"
)

func main() {
	router := plugo.New()

	router.Get("/", hello_world)

	fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func hello_world(w http.ResponseWriter, r *http.Request) {
	conn := plugo.NewConnection(w, r)

	conn.String(http.StatusOK, "Hello, Plugo World!")
}

```
