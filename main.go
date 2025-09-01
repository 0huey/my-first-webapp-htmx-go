package main

import(
	"net/http"
	"html/template"
	"log"
	"reflect")

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

	http.Handle("GET /static/",
		http.StripPrefix("/static", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", HandleRoot)

	http.HandleFunc("POST /count", HandleCount)

	http.HandleFunc("/contacts", HandleContact)

	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func HandleRoot(w http.ResponseWriter, req *http.Request) {
	Render(w, req, "index", context)
}

func HandleCount(w http.ResponseWriter, req *http.Request) {
	context.Count++
	Render(w, req, "count", context)
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
		RenderError(w, req, "email form", temp_context, http.StatusConflict)
		return
	}

	new := NewContact(name, email)
	context.Contacts.AddContact(new)
	Render(w, req, "oob-contact", new)
	Render(w, req, "email form", context)
}

func Render(w http.ResponseWriter, req *http.Request, block string, context any) {
	t := template.Must(template.ParseGlob("templates/*.html"))

	w.Header().Add("Cache-Control", "no-cache")

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Println(err)
	}

	private := reflect.ValueOf(w).Elem()

	log.Println(req.RemoteAddr,
		req.Method,
		req.RequestURI,
		private.FieldByName("status"),
		private.FieldByName("written"),
	)
}

func RenderError(w http.ResponseWriter, req *http.Request, block string, context any, code int) {
	w.WriteHeader(code)
	w.Header().Add("Cache-Control", "no-cache")

	t := template.Must(template.ParseGlob("templates/*.html"))

	w.Header().Add("Cache-Control", "no-cache")

	err := t.ExecuteTemplate(w, block, context)

	if err != nil {
		log.Println(err)
	}

	private := reflect.ValueOf(w).Elem()

	log.Println(req.RemoteAddr,
		req.Method,
		req.RequestURI,
		code,
		private.FieldByName("written"),
	)
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
