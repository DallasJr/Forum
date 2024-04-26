package main

import (
	"Forum/src"
	"Forum/src/handlers"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

func main() {
	src.SetupDatabase()
	handlers.SetupHandlers()

	fmt.Println("http://localhost/")
	err := http.ListenAndServe("", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
}
