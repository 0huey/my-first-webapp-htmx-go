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
type ContactSlice []Contact

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type ClientContext struct {
	Count int
	Contacts ContactSlice
	FormErrors FormData
}

var context ClientContext

func main() {
	// some initial data
	context.Contacts.AddContact(NewContact("john smith", "js@example.com"))
	context.Contacts.AddContact(NewContact("person", "person@example.com"))

	http.HandleFunc("/", HandleRoot)

	http.HandleFunc("POST /count", HandleCount)

	http.HandleFunc("POST /contacts", HandleContact)

	http.ListenAndServe(":8080", nil)
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	Render(w, "index", context)
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	context.Count++
	Render(w, "count", context)
}

func HandleContact(w http.ResponseWriter, req *http.Request) {
	name  := req.PostFormValue("name")
	email := req.PostFormValue("email")

	if len(name) == 0 || len(email) == 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if context.Contacts.EmailExists(email) {
		temp_context := context
		temp_context.FormErrors = newFormData()
		temp_context.FormErrors.Values["name"] = name
		temp_context.FormErrors.Values["email"] = email
		temp_context.FormErrors.Errors["message"] = "that email address already exists retard"
		RenderError(w, "email form", temp_context, http.StatusConflict)
		return
	}

	new := NewContact(name, email)
	context.Contacts.AddContact(new)
	Render(w, "oob-contact", new)
	Render(w, "email form", context)
}

func Render(w http.ResponseWriter, block string, context any) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	w.Header().Add("Cache-Control", "no-cache")

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		fmt.Println(err)
	}
}

func RenderError(w http.ResponseWriter, block string, context any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Cache-Control", "no-cache")
	Render(w, block, context)
}

func NewContact(name string, email string) Contact {
	return Contact {Name: name, Email: email}
}

func (c *ContactSlice) AddContact(new Contact) {
	*c = append(*c, new)
}

func (c ContactSlice) EmailExists(email string) bool {
	for _, contact := range c {
		if contact.Email == email {
			return true
		}
	}
	return false
}
