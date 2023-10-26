/*
Plugo is a package to easily write a HTTP request router for your backend.

You can write the following example of hello world to test:

	func main() {
		router := plugo.New()

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			conn := plugo.NewConnection(w, r)

			conn.String(http.StatusOK, "Hello, Plugo World!")
		})

		fmt.Println("Server running at http://localhost:8080/ - Press CTRL+C to exit")
		log.Fatal(http.ListenAndServe(":8080", router))
	}
*/
package plugo
