package main

import(
	"net/http"
	"html/template"
)

type Count struct {
	Count int
}

var count Count

func main() {
	count = Count{Count: 0}

	http.HandleFunc("/", HandleRoot)

	http.ListenAndServe(":8080", nil)
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	count.Count++
	Render(w, "templates/index.html", count)
}

func Render(w http.ResponseWriter, filename string, context any) {
	template.Must(template.ParseFiles(filename)).Execute(w, context)
}
