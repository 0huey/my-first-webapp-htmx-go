package main

import(
	"net/http"
	"html/template"
	"fmt"
)

type Count struct {
	Count int
}

var count Count

func main() {
	count = Count{Count: 0}

	http.HandleFunc("/", HandleRoot)

	http.HandleFunc("/count", HandleCount)

	http.ListenAndServe(":8080", nil)
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	Render(w, "index", count)
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		count.Count++
		Render(w, "count", count)
	}
}

func Render(w http.ResponseWriter, block string, context any) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		fmt.Println(err)
	}
}
