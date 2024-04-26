package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func main() {
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src"))))
	http.HandleFunc("/", index)
	//http.HandleFunc("/test", test)
	fmt.Println("http://localhost/")
	http.ListenAndServe("", nil)
}

type exportData struct {
}

var ExportData exportData

func index(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon.ico") {
		return
	}
	tmpl := template.Must(template.ParseFiles("src/templates/index.html"))
	tmpl.Execute(w, ExportData)
}
