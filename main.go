package main

import(
	"net/http"
	"html/template"
)

type Count struct {
	Count int
}

func main() {
	count := Count{Count: 0}

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request) {
		count.Count++
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, count)
	})

	http.ListenAndServe(":8080", nil)
}
