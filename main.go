package main

import(
	"net/http"
	"html/template"
	"fmt"
)

type Contact struct {
	Name string
	Email string
}

type IndexContext struct {
	Count int
	Contacts []Contact
}

var index_context IndexContext

func main() {
	// some initial data
	NewContact("john smith", "js@example.com")
	NewContact("person", "person@example.com")

	http.HandleFunc("/", HandleRoot)

	http.HandleFunc("POST /count", HandleCount)

	http.HandleFunc("POST /contacts", HandleContact)

	http.ListenAndServe(":8080", nil)
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	Render(w, "index", index_context)
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	index_context.Count++
	Render(w, "count", index_context)
}

func HandleContact(w http.ResponseWriter, req *http.Request) {
	name :=  req.PostFormValue("name")
	email := req.PostFormValue("email")

	if len(name) == 0 || len(email) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	NewContact(name, email)

	Render(w, "contacts", index_context)
}

func Render(w http.ResponseWriter, block string, context any) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		fmt.Println(err)
	}
}

func NewContact(name string, email string) {
	c := Contact {Name: name, Email: email}
	index_context.Contacts = append(index_context.Contacts, c)
}
